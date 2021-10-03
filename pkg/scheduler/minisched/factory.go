/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mini_kube_scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	internalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"
	internalqueue "k8s.io/kubernetes/pkg/scheduler/internal/queue"
	"k8s.io/kubernetes/pkg/scheduler/profile"
)

// Binder knows how to write a binding.
type Binder interface {
	Bind(binding *v1.Binding) error
}

// Configurator defines I/O, caching, and other functionality needed to
// construct a new scheduler.
type Configurator struct {
	client     clientset.Interface
	kubeConfig *restclient.Config

	recorderFactory profile.RecorderFactory

	informerFactory informers.SharedInformerFactory

	// Close this to stop all reflectors
	StopEverything <-chan struct{}

	schedulerCache Cache

	componentConfigVersion string

	// Always check all predicates even if the middle of one predicate fails.
	alwaysCheckAllPredicates bool

	// percentageOfNodesToScore specifies percentage of all nodes to score in each scheduling cycle.
	percentageOfNodesToScore int32

	podInitialBackoffSeconds int64

	podMaxBackoffSeconds int64

	profiles          []schedulerapi.KubeSchedulerProfile
	registry          frameworkruntime.Registry
	nodeInfoSnapshot  *internalcache.Snapshot
	extenders         []schedulerapi.Extender
	frameworkCapturer FrameworkCapturer
	parallellism      int32
	// A "cluster event" -> "plugin names" map.
	clusterEventMap map[framework.ClusterEvent]sets.String
}

// create a scheduler from a set of registered plugins.
func (c *Configurator) create() (*Scheduler, error) {
	if len(c.extenders) != 0 {
		// We don't support extenders on mini-kube-scheduler.
	}

	// The nominator will be passed all the way to framework instantiation.
	nominator := internalqueue.NewPodNominator(c.informerFactory.Core().V1().Pods().Lister())
	profiles, err := profile.NewMap(c.profiles, c.registry, c.recorderFactory,
		frameworkruntime.WithComponentConfigVersion(c.componentConfigVersion),
		frameworkruntime.WithClientSet(c.client),
		frameworkruntime.WithKubeConfig(c.kubeConfig),
		frameworkruntime.WithInformerFactory(c.informerFactory),
		frameworkruntime.WithSnapshotSharedLister(c.nodeInfoSnapshot),
		frameworkruntime.WithRunAllFilters(c.alwaysCheckAllPredicates),
		frameworkruntime.WithPodNominator(nominator),
		frameworkruntime.WithCaptureProfile(frameworkruntime.CaptureProfile(c.frameworkCapturer)),
		frameworkruntime.WithClusterEventMap(c.clusterEventMap),
		frameworkruntime.WithParallelism(int(c.parallellism)),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing profiles: %v", err)
	}
	if len(profiles) != 1 {
		// mini-kube-scheduler support only one profile.
		return nil, errors.New("at least one profile is required")
	}
	profile :=c.profiles[0]

	lessFn := profiles[profile.SchedulerName].QueueSortFunc()

	podQueue := internalqueue.NewSchedulingQueue(
		lessFn,
		c.informerFactory,
		internalqueue.WithPodInitialBackoffDuration(time.Duration(c.podInitialBackoffSeconds)*time.Second),
		internalqueue.WithPodMaxBackoffDuration(time.Duration(c.podMaxBackoffSeconds)*time.Second),
		internalqueue.WithPodNominator(nominator),
		internalqueue.WithClusterEventMap(c.clusterEventMap),
	)

	return &Scheduler{
		SchedulerCache:  c.schedulerCache,
		Error:           MakeDefaultErrorFunc(c.client, c.informerFactory.Core().V1().Pods().Lister(), podQueue, c.schedulerCache),
		SchedulingQueue: podQueue,
		preFilterPlugins : profile.
		filterPlugins   :
		scorePlugins   :
		preScorePlugins:
	}, nil
}

// ErrNoNodesAvailable is used to describe the error that no nodes available to schedule pods.
var ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule pods")

// MakeDefaultErrorFunc construct a function to handle pod scheduler error
func MakeDefaultErrorFunc(client clientset.Interface, podLister corelisters.PodLister, podQueue internalqueue.SchedulingQueue, schedulerCache internalcache.Cache) func(*framework.QueuedPodInfo, error) {
	return func(podInfo *framework.QueuedPodInfo, err error) {
		pod := podInfo.Pod
		if err == ErrNoNodesAvailable {
			klog.V(2).InfoS("Unable to schedule pod; no nodes are registered to the cluster; waiting", "pod", klog.KObj(pod))
		} else if fitError, ok := err.(*framework.FitError); ok {
			// Inject UnschedulablePlugins to PodInfo, which will be used later for moving Pods between queues efficiently.
			podInfo.UnschedulablePlugins = fitError.Diagnosis.UnschedulablePlugins
			klog.V(2).InfoS("Unable to schedule pod; no fit; waiting", "pod", klog.KObj(pod), "err", err)
		} else if apierrors.IsNotFound(err) {
			klog.V(2).InfoS("Unable to schedule pod, possibly due to node not found; waiting", "pod", klog.KObj(pod), "err", err)
			if errStatus, ok := err.(apierrors.APIStatus); ok && errStatus.Status().Details.Kind == "node" {
				nodeName := errStatus.Status().Details.Name
				// when node is not found, We do not remove the node right away. Trying again to get
				// the node and if the node is still not found, then remove it from the scheduler cache.
				_, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
				if err != nil && apierrors.IsNotFound(err) {
					node := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}
					if err := schedulerCache.RemoveNode(&node); err != nil {
						klog.V(4).InfoS("Node is not found; failed to remove it from the cache", "node", node.Name)
					}
				}
			}
		} else {
			klog.ErrorS(err, "Error scheduling pod; retrying", "pod", klog.KObj(pod))
		}

		// Check if the Pod exists in informer cache.
		cachedPod, err := podLister.Pods(pod.Namespace).Get(pod.Name)
		if err != nil {
			klog.InfoS("Pod doesn't exist in informer cache", "pod", klog.KObj(pod), "err", err)
			return
		}

		// In the case of extender, the pod may have been bound successfully, but timed out returning its response to the scheduler.
		// It could result in the live version to carry .spec.nodeName, and that's inconsistent with the internal-queued version.
		if len(cachedPod.Spec.NodeName) != 0 {
			klog.InfoS("Pod has been assigned to node. Abort adding it back to queue.", "pod", klog.KObj(pod), "node", cachedPod.Spec.NodeName)
			return
		}

		// As <cachedPod> is from SharedInformer, we need to do a DeepCopy() here.
		podInfo.PodInfo = framework.NewPodInfo(cachedPod.DeepCopy())
		if err := podQueue.AddUnschedulableIfNotPresent(podInfo, podQueue.SchedulingCycle()); err != nil {
			klog.ErrorS(err, "Error occurred")
		}
	}
}
