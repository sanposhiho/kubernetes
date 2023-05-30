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

package scheduler

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	internalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"
	internalqueue "k8s.io/kubernetes/pkg/scheduler/internal/queue"
	"k8s.io/kubernetes/pkg/scheduler/profile"
	st "k8s.io/kubernetes/pkg/scheduler/testing"
	testingclock "k8s.io/utils/clock/testing"
	"k8s.io/utils/pointer"
)

func TestSchedulerCreation(t *testing.T) {
	invalidRegistry := map[string]frameworkruntime.PluginFactory{
		defaultbinder.Name: defaultbinder.New,
	}
	validRegistry := map[string]frameworkruntime.PluginFactory{
		"Foo": defaultbinder.New,
	}
	cases := []struct {
		name          string
		opts          []Option
		wantErr       string
		wantProfiles  []string
		wantExtenders []string
	}{
		{
			name: "valid out-of-tree registry",
			opts: []Option{
				WithFrameworkOutOfTreeRegistry(validRegistry),
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantProfiles: []string{"default-scheduler"},
		},
		{
			name: "repeated plugin name in out-of-tree plugin",
			opts: []Option{
				WithFrameworkOutOfTreeRegistry(invalidRegistry),
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantProfiles: []string{"default-scheduler"},
			wantErr:      "a plugin named DefaultBinder already exists",
		},
		{
			name: "multiple profiles",
			opts: []Option{
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "bar",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantProfiles: []string{"bar", "foo"},
		},
		{
			name: "Repeated profiles",
			opts: []Option{
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "bar",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantErr: "duplicate profile with scheduler name \"foo\"",
		},
		{
			name: "With extenders",
			opts: []Option{
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				),
				WithExtenders(
					schedulerapi.Extender{
						URLPrefix: "http://extender.kube-system/",
					},
				),
			},
			wantProfiles:  []string{"default-scheduler"},
			wantExtenders: []string{"http://extender.kube-system/"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, 0)

			eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: client.EventsV1()})

			_, ctx := ktesting.NewTestContext(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			s, err := New(
				ctx,
				client,
				informerFactory,
				nil,
				profile.NewRecorderFactory(eventBroadcaster),
				tc.opts...,
			)

			// Errors
			if len(tc.wantErr) != 0 {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("got error %q, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to create scheduler: %v", err)
			}

			// Profiles
			profiles := make([]string, 0, len(s.Profiles))
			for name := range s.Profiles {
				profiles = append(profiles, name)
			}
			sort.Strings(profiles)
			if diff := cmp.Diff(tc.wantProfiles, profiles); diff != "" {
				t.Errorf("unexpected profiles (-want, +got):\n%s", diff)
			}

			// Extenders
			if len(tc.wantExtenders) != 0 {
				// Scheduler.Extenders
				extenders := make([]string, 0, len(s.Extenders))
				for _, e := range s.Extenders {
					extenders = append(extenders, e.Name())
				}
				if diff := cmp.Diff(tc.wantExtenders, extenders); diff != "" {
					t.Errorf("unexpected extenders (-want, +got):\n%s", diff)
				}

				// framework.Handle.Extenders()
				for _, p := range s.Profiles {
					extenders := make([]string, 0, len(p.Extenders()))
					for _, e := range p.Extenders() {
						extenders = append(extenders, e.Name())
					}
					if diff := cmp.Diff(tc.wantExtenders, extenders); diff != "" {
						t.Errorf("unexpected extenders (-want, +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestFailureHandler(t *testing.T) {
	testPod := st.MakePod().Name("test-pod").Namespace(v1.NamespaceDefault).Obj()
	testPodUpdated := testPod.DeepCopy()
	testPodUpdated.Labels = map[string]string{"foo": ""}

	tests := []struct {
		name                       string
		podUpdatedDuringScheduling bool // pod is updated during a scheduling cycle
		podDeletedDuringScheduling bool // pod is deleted during a scheduling cycle
		expect                     *v1.Pod
	}{
		{
			name:                       "pod is updated during a scheduling cycle",
			podUpdatedDuringScheduling: true,
			expect:                     testPodUpdated,
		},
		{
			name:   "pod is not updated during a scheduling cycle",
			expect: testPod,
		},
		{
			name:                       "pod is deleted during a scheduling cycle",
			podDeletedDuringScheduling: true,
			expect:                     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx := ktesting.NewTestContext(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			client := fake.NewSimpleClientset(&v1.PodList{Items: []v1.Pod{*testPod}})
			informerFactory := informers.NewSharedInformerFactory(client, 0)
			podInformer := informerFactory.Core().V1().Pods()
			// Need to add/update/delete testPod to the store.
			podInformer.Informer().GetStore().Add(testPod)

			queue := internalqueue.NewPriorityQueue(nil, informerFactory, internalqueue.WithClock(testingclock.NewFakeClock(time.Now())))
			schedulerCache := internalcache.New(ctx, 30*time.Second)

			queue.Add(logger, testPod)
			queue.Pop()

			if tt.podUpdatedDuringScheduling {
				podInformer.Informer().GetStore().Update(testPodUpdated)
				queue.Update(logger, testPod, testPodUpdated)
			}
			if tt.podDeletedDuringScheduling {
				podInformer.Informer().GetStore().Delete(testPod)
				queue.Delete(testPod)
			}

			s, fwk, err := initScheduler(ctx, schedulerCache, queue, client, informerFactory)
			if err != nil {
				t.Fatal(err)
			}

			testPodInfo := &framework.QueuedPodInfo{PodInfo: mustNewPodInfo(t, testPod)}
			s.FailureHandler(ctx, fwk, testPodInfo, framework.NewStatus(framework.Unschedulable), nil, nil, time.Now())

			var got *v1.Pod
			if tt.podUpdatedDuringScheduling {
				head, e := queue.Pop()
				if e != nil {
					t.Fatalf("Cannot pop pod from the activeQ: %v", e)
				}
				got = head.Pod
			} else {
				got = getPodFromPriorityQueue(queue, testPod)
			}

			if diff := cmp.Diff(tt.expect, got); diff != "" {
				t.Errorf("Unexpected pod (-want, +got): %s", diff)
			}
		})
	}
}

func TestFailureHandler_NodeNotFound(t *testing.T) {
	nodeFoo := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
	nodeBar := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "bar"}}
	testPod := st.MakePod().Name("test-pod").Namespace(v1.NamespaceDefault).Obj()
	tests := []struct {
		name             string
		nodes            []v1.Node
		nodeNameToDelete string
		injectErr        error
		expectNodeNames  sets.Set[string]
	}{
		{
			name:             "node is deleted during a scheduling cycle",
			nodes:            []v1.Node{*nodeFoo, *nodeBar},
			nodeNameToDelete: "foo",
			injectErr:        apierrors.NewNotFound(v1.Resource("node"), nodeFoo.Name),
			expectNodeNames:  sets.New("bar"),
		},
		{
			name:            "node is not deleted but NodeNotFound is received incorrectly",
			nodes:           []v1.Node{*nodeFoo, *nodeBar},
			injectErr:       apierrors.NewNotFound(v1.Resource("node"), nodeFoo.Name),
			expectNodeNames: sets.New("foo", "bar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx := ktesting.NewTestContext(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			client := fake.NewSimpleClientset(&v1.PodList{Items: []v1.Pod{*testPod}}, &v1.NodeList{Items: tt.nodes})
			informerFactory := informers.NewSharedInformerFactory(client, 0)
			podInformer := informerFactory.Core().V1().Pods()
			// Need to add testPod to the store.
			podInformer.Informer().GetStore().Add(testPod)

			queue := internalqueue.NewPriorityQueue(nil, informerFactory, internalqueue.WithClock(testingclock.NewFakeClock(time.Now())))
			schedulerCache := internalcache.New(ctx, 30*time.Second)

			for i := range tt.nodes {
				node := tt.nodes[i]
				// Add node to schedulerCache no matter it's deleted in API server or not.
				schedulerCache.AddNode(logger, &node)
				if node.Name == tt.nodeNameToDelete {
					client.CoreV1().Nodes().Delete(ctx, node.Name, metav1.DeleteOptions{})
				}
			}

			s, fwk, err := initScheduler(ctx, schedulerCache, queue, client, informerFactory)
			if err != nil {
				t.Fatal(err)
			}

			testPodInfo := &framework.QueuedPodInfo{PodInfo: mustNewPodInfo(t, testPod)}
			s.FailureHandler(ctx, fwk, testPodInfo, framework.NewStatus(framework.Unschedulable).WithError(tt.injectErr), nil, nil, time.Now())

			gotNodes := schedulerCache.Dump().Nodes
			gotNodeNames := sets.New[string]()
			for _, nodeInfo := range gotNodes {
				gotNodeNames.Insert(nodeInfo.Node().Name)
			}
			if diff := cmp.Diff(tt.expectNodeNames, gotNodeNames); diff != "" {
				t.Errorf("Unexpected nodes (-want, +got): %s", diff)
			}
		})
	}
}

func TestFailureHandler_PodAlreadyBound(t *testing.T) {
	logger, ctx := ktesting.NewTestContext(t)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	nodeFoo := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
	testPod := st.MakePod().Name("test-pod").Namespace(v1.NamespaceDefault).Node("foo").Obj()

	client := fake.NewSimpleClientset(&v1.PodList{Items: []v1.Pod{*testPod}}, &v1.NodeList{Items: []v1.Node{nodeFoo}})
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	podInformer := informerFactory.Core().V1().Pods()
	// Need to add testPod to the store.
	podInformer.Informer().GetStore().Add(testPod)

	queue := internalqueue.NewPriorityQueue(nil, informerFactory, internalqueue.WithClock(testingclock.NewFakeClock(time.Now())))
	schedulerCache := internalcache.New(ctx, 30*time.Second)

	// Add node to schedulerCache no matter it's deleted in API server or not.
	schedulerCache.AddNode(logger, &nodeFoo)

	s, fwk, err := initScheduler(ctx, schedulerCache, queue, client, informerFactory)
	if err != nil {
		t.Fatal(err)
	}

	testPodInfo := &framework.QueuedPodInfo{PodInfo: mustNewPodInfo(t, testPod)}
	s.FailureHandler(ctx, fwk, testPodInfo, framework.NewStatus(framework.Unschedulable).WithError(fmt.Errorf("binding rejected: timeout")), nil, nil, time.Now())

	pod := getPodFromPriorityQueue(queue, testPod)
	if pod != nil {
		t.Fatalf("Unexpected pod: %v should not be in PriorityQueue when the NodeName of pod is not empty", pod.Name)
	}
}

// TestWithPercentageOfNodesToScore tests scheduler's PercentageOfNodesToScore is set correctly.
func TestWithPercentageOfNodesToScore(t *testing.T) {
	tests := []struct {
		name                           string
		percentageOfNodesToScoreConfig *int32
		wantedPercentageOfNodesToScore int32
	}{
		{
			name:                           "percentageOfNodesScore is nil",
			percentageOfNodesToScoreConfig: nil,
			wantedPercentageOfNodesToScore: schedulerapi.DefaultPercentageOfNodesToScore,
		},
		{
			name:                           "percentageOfNodesScore is not nil",
			percentageOfNodesToScoreConfig: pointer.Int32(10),
			wantedPercentageOfNodesToScore: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, 0)
			eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: client.EventsV1()})
			_, ctx := ktesting.NewTestContext(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			sched, err := New(
				ctx,
				client,
				informerFactory,
				nil,
				profile.NewRecorderFactory(eventBroadcaster),
				WithPercentageOfNodesToScore(tt.percentageOfNodesToScoreConfig),
			)
			if err != nil {
				t.Fatalf("Failed to create scheduler: %v", err)
			}
			if sched.percentageOfNodesToScore != tt.wantedPercentageOfNodesToScore {
				t.Errorf("scheduler.percercentageOfNodesToScore = %v, want %v", sched.percentageOfNodesToScore, tt.wantedPercentageOfNodesToScore)
			}
		})
	}
}

// getPodFromPriorityQueue is the function used in the TestDefaultErrorFunc test to get
// the specific pod from the given priority queue. It returns the found pod in the priority queue.
func getPodFromPriorityQueue(queue *internalqueue.PriorityQueue, pod *v1.Pod) *v1.Pod {
	podList, _ := queue.PendingPods()
	if len(podList) == 0 {
		return nil
	}

	queryPodKey, err := cache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		return nil
	}

	for _, foundPod := range podList {
		foundPodKey, err := cache.MetaNamespaceKeyFunc(foundPod)
		if err != nil {
			return nil
		}

		if foundPodKey == queryPodKey {
			return foundPod
		}
	}

	return nil
}

func initScheduler(ctx context.Context, cache internalcache.Cache, queue internalqueue.SchedulingQueue,
	client kubernetes.Interface, informerFactory informers.SharedInformerFactory) (*Scheduler, framework.Framework, error) {
	logger := klog.FromContext(ctx)
	registerPluginFuncs := []st.RegisterPluginFunc{
		st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
		st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
	}
	eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: client.EventsV1()})
	fwk, err := st.NewFramework(ctx,
		registerPluginFuncs,
		testSchedulerName,
		frameworkruntime.WithClientSet(client),
		frameworkruntime.WithInformerFactory(informerFactory),
		frameworkruntime.WithEventRecorder(eventBroadcaster.NewRecorder(scheme.Scheme, testSchedulerName)),
	)
	if err != nil {
		return nil, nil, err
	}

	s := &Scheduler{
		Cache:           cache,
		client:          client,
		StopEverything:  ctx.Done(),
		SchedulingQueue: queue,
		Profiles:        profile.Map{testSchedulerName: fwk},
		logger:          logger,
	}
	s.applyDefaultHandlers()

	return s, fwk, nil
}

func TestInitPluginsWithIndexers(t *testing.T) {
	tests := []struct {
		name string
		// the plugin registration ordering must not matter, being map traversal random
		entrypoints map[string]frameworkruntime.PluginFactory
		wantErr     string
	}{
		{
			name: "register indexer, no conflicts",
			entrypoints: map[string]frameworkruntime.PluginFactory{
				"AddIndexer": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
					podInformer := handle.SharedInformerFactory().Core().V1().Pods()
					err := podInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
						"nodeName": indexByPodSpecNodeName,
					})
					return &TestPlugin{name: "AddIndexer"}, err
				},
			},
		},
		{
			name: "register the same indexer name multiple times, conflict",
			// order of registration doesn't matter
			entrypoints: map[string]frameworkruntime.PluginFactory{
				"AddIndexer1": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
					podInformer := handle.SharedInformerFactory().Core().V1().Pods()
					err := podInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
						"nodeName": indexByPodSpecNodeName,
					})
					return &TestPlugin{name: "AddIndexer1"}, err
				},
				"AddIndexer2": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
					podInformer := handle.SharedInformerFactory().Core().V1().Pods()
					err := podInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
						"nodeName": indexByPodAnnotationNodeName,
					})
					return &TestPlugin{name: "AddIndexer1"}, err
				},
			},
			wantErr: "indexer conflict",
		},
		{
			name: "register the same indexer body with different names, no conflicts",
			// order of registration doesn't matter
			entrypoints: map[string]frameworkruntime.PluginFactory{
				"AddIndexer1": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
					podInformer := handle.SharedInformerFactory().Core().V1().Pods()
					err := podInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
						"nodeName1": indexByPodSpecNodeName,
					})
					return &TestPlugin{name: "AddIndexer1"}, err
				},
				"AddIndexer2": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
					podInformer := handle.SharedInformerFactory().Core().V1().Pods()
					err := podInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
						"nodeName2": indexByPodAnnotationNodeName,
					})
					return &TestPlugin{name: "AddIndexer2"}, err
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeInformerFactory := NewInformerFactory(&fake.Clientset{}, 0*time.Second)

			var registerPluginFuncs []st.RegisterPluginFunc
			for name, entrypoint := range tt.entrypoints {
				registerPluginFuncs = append(registerPluginFuncs,
					// anything supported by TestPlugin is fine
					st.RegisterFilterPlugin(name, entrypoint),
				)
			}
			// we always need this
			registerPluginFuncs = append(registerPluginFuncs,
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			)
			_, ctx := ktesting.NewTestContext(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			_, err := st.NewFramework(ctx, registerPluginFuncs, "test", frameworkruntime.WithInformerFactory(fakeInformerFactory))

			if len(tt.wantErr) > 0 {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("got error %q, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to create scheduler: %v", err)
			}
		})
	}
}

func indexByPodSpecNodeName(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return []string{}, nil
	}
	if len(pod.Spec.NodeName) == 0 {
		return []string{}, nil
	}
	return []string{pod.Spec.NodeName}, nil
}

func indexByPodAnnotationNodeName(obj interface{}) ([]string, error) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return []string{}, nil
	}
	if len(pod.Annotations) == 0 {
		return []string{}, nil
	}
	nodeName, ok := pod.Annotations["node-name"]
	if !ok {
		return []string{}, nil
	}
	return []string{nodeName}, nil
}

const (
	fakeNoop        = "fakeNoop"
	fakeNode        = "fakeNode"
	fakePod         = "fakePod"
	queueSortPlugin = "no-op-queue-sort-plugin"
	bindPlugin      = "bind-plugin"
)

func Test_buildQueueingHintMap(t *testing.T) {
	tests := []struct {
		name    string
		plugins []framework.Plugin
		want    map[framework.ClusterEvent]sets.Set[string]
	}{
		{
			name:    "no-op plugin",
			plugins: []framework.Plugin{&fakeNoopPlugin{}},
			want: map[framework.ClusterEvent]sets.Set[string]{
				{Resource: framework.Pod, ActionType: framework.All}:                   sets.New(fakeNoop, bindPlugin, queueSortPlugin),
				{Resource: framework.Node, ActionType: framework.All}:                  sets.New(fakeNoop, bindPlugin, queueSortPlugin),
				{Resource: framework.CSINode, ActionType: framework.All}:               sets.New(fakeNoop, bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolume, ActionType: framework.All}:      sets.New(fakeNoop, bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}: sets.New(fakeNoop, bindPlugin, queueSortPlugin),
				{Resource: framework.StorageClass, ActionType: framework.All}:          sets.New(fakeNoop, bindPlugin, queueSortPlugin),
			},
		},
		{
			name:    "node plugin",
			plugins: []framework.Plugin{&fakeNodePlugin{}},
			want: map[framework.ClusterEvent]sets.Set[string]{
				{Resource: framework.Pod, ActionType: framework.All}:                           sets.New(fakeNode, bindPlugin, queueSortPlugin),
				{Resource: framework.Node, ActionType: framework.Delete}:                       sets.New(fakeNode),
				{Resource: framework.Node, ActionType: framework.All}:                          sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.CSINode, ActionType: framework.Update | framework.Delete}: sets.New(fakeNode),
				{Resource: framework.CSINode, ActionType: framework.All}:                       sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolume, ActionType: framework.All}:              sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}:         sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.StorageClass, ActionType: framework.All}:                  sets.New(bindPlugin, queueSortPlugin),
			},
		},
		{
			name:    "pod plugin",
			plugins: []framework.Plugin{&fakePodPlugin{}},
			want: map[framework.ClusterEvent]sets.Set[string]{
				{Resource: framework.Pod, ActionType: framework.All}:                      sets.New(fakePod, bindPlugin, queueSortPlugin),
				{Resource: framework.Node, ActionType: framework.Add | framework.Delete}:  sets.New(fakePod),
				{Resource: framework.Node, ActionType: framework.All}:                     sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.Delete}: sets.New(fakePod),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}:    sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.CSINode, ActionType: framework.All}:                  sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolume, ActionType: framework.All}:         sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.StorageClass, ActionType: framework.All}:             sets.New(bindPlugin, queueSortPlugin),
			},
		},
		{
			name:    "node and pod plugin",
			plugins: []framework.Plugin{&fakeNodePlugin{}, &fakePodPlugin{}},
			want: map[framework.ClusterEvent]sets.Set[string]{
				{Resource: framework.Node, ActionType: framework.Delete}:                       sets.New(fakeNode),
				{Resource: framework.Node, ActionType: framework.Add | framework.Delete}:       sets.New(fakePod),
				{Resource: framework.Pod, ActionType: framework.All}:                           sets.New(fakeNode, fakePod, bindPlugin, queueSortPlugin),
				{Resource: framework.CSINode, ActionType: framework.Update | framework.Delete}: sets.New(fakeNode),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.Delete}:      sets.New(fakePod),
				{Resource: framework.Node, ActionType: framework.All}:                          sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.CSINode, ActionType: framework.All}:                       sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolume, ActionType: framework.All}:              sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}:         sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.StorageClass, ActionType: framework.All}:                  sets.New(bindPlugin, queueSortPlugin),
			},
		},
		{
			name:    "no-op runtime plugin",
			plugins: []framework.Plugin{&fakeNoopRuntimePlugin{}},
			want: map[framework.ClusterEvent]sets.Set[string]{
				{Resource: framework.Pod, ActionType: framework.All}:                   sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.Node, ActionType: framework.All}:                  sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.CSINode, ActionType: framework.All}:               sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolume, ActionType: framework.All}:      sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}: sets.New(bindPlugin, queueSortPlugin),
				{Resource: framework.StorageClass, ActionType: framework.All}:          sets.New(bindPlugin, queueSortPlugin),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := frameworkruntime.Registry{}
			cfgPls := &schedulerapi.Plugins{}
			for _, pl := range tt.plugins {
				tmpPl := pl
				if err := registry.Register(pl.Name(), func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
					return tmpPl, nil
				}); err != nil {
					t.Fatalf("fail to register filter plugin (%s)", pl.Name())
				}
				cfgPls.Filter.Enabled = append(cfgPls.Filter.Enabled, schedulerapi.Plugin{Name: pl.Name()})
			}

			got := make(map[framework.ClusterEvent]sets.Set[string])
			profile := schedulerapi.KubeSchedulerProfile{Plugins: cfgPls}
			stopCh := make(chan struct{})
			defer close(stopCh)
			framework, err := newFrameworkWithQueueSortAndBind(registry, profile, stopCh)
			if err != nil {
				t.Fatal(err)
			}

			buildQueueingHintMap(framework.EnqueueExtensions(), got)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Unexpected eventToPlugin map (-want,+got):%s", diff)
			}
		})
	}
}

func newFrameworkWithQueueSortAndBind(r frameworkruntime.Registry, profile schedulerapi.KubeSchedulerProfile, stopCh <-chan struct{}) (framework.Framework, error) {
	if _, ok := r[queueSortPlugin]; !ok {
		r[queueSortPlugin] = newQueueSortPlugin
	}
	if _, ok := r[bindPlugin]; !ok {
		r[bindPlugin] = newBindPlugin
	}

	if len(profile.Plugins.QueueSort.Enabled) == 0 {
		profile.Plugins.QueueSort.Enabled = append(profile.Plugins.QueueSort.Enabled, schedulerapi.Plugin{Name: queueSortPlugin})
	}
	if len(profile.Plugins.Bind.Enabled) == 0 {
		profile.Plugins.Bind.Enabled = append(profile.Plugins.Bind.Enabled, schedulerapi.Plugin{Name: bindPlugin})
	}
	return frameworkruntime.NewFramework(context.Background(), r, &profile)
}

var _ framework.QueueSortPlugin = &TestQueueSortPlugin{}

func newQueueSortPlugin(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &TestQueueSortPlugin{}, nil
}

// TestQueueSortPlugin is a no-op implementation for QueueSort extension point.
type TestQueueSortPlugin struct{}

func (pl *TestQueueSortPlugin) Name() string {
	return queueSortPlugin
}

func (pl *TestQueueSortPlugin) Less(_, _ *framework.QueuedPodInfo) bool {
	return false
}

var _ framework.BindPlugin = &TestBindPlugin{}

func newBindPlugin(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &TestBindPlugin{}, nil
}

// TestBindPlugin is a no-op implementation for Bind extension point.
type TestBindPlugin struct{}

func (t TestBindPlugin) Name() string {
	return bindPlugin
}

func (t TestBindPlugin) Bind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	return nil
}

// fakeNoopPlugin doesn't implement interface framework.EnqueueExtensions.
type fakeNoopPlugin struct{}

func (*fakeNoopPlugin) Name() string { return fakeNoop }

func (*fakeNoopPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return nil
}

type fakeNodePlugin struct{}

func (*fakeNodePlugin) Name() string { return fakeNode }

func (*fakeNodePlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return nil
}

func (*fakeNodePlugin) EventsToRegister() []framework.ClusterEventWithHint {
	return []framework.ClusterEventWithHint{
		{Event: framework.ClusterEvent{Resource: framework.Pod, ActionType: framework.All}},
		{Event: framework.ClusterEvent{Resource: framework.Node, ActionType: framework.Delete}},
		{Event: framework.ClusterEvent{Resource: framework.CSINode, ActionType: framework.Update | framework.Delete}},
	}
}

type fakePodPlugin struct{}

func (*fakePodPlugin) Name() string { return fakePod }

func (*fakePodPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return nil
}

func (*fakePodPlugin) EventsToRegister() []framework.ClusterEventWithHint {
	return []framework.ClusterEventWithHint{
		{Event: framework.ClusterEvent{Resource: framework.Pod, ActionType: framework.All}},
		{Event: framework.ClusterEvent{Resource: framework.Node, ActionType: framework.Add | framework.Delete}},
		{Event: framework.ClusterEvent{Resource: framework.PersistentVolumeClaim, ActionType: framework.Delete}},
	}
}

// fakeNoopRuntimePlugin implement interface framework.EnqueueExtensions, but returns nil
// at runtime. This can simulate a plugin registered at scheduler setup, but does nothing
// due to some disabled feature gate.
type fakeNoopRuntimePlugin struct{}

func (*fakeNoopRuntimePlugin) Name() string { return "fakeNoopRuntime" }

func (*fakeNoopRuntimePlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return nil
}

func (*fakeNoopRuntimePlugin) EventsToRegister() []framework.ClusterEventWithHint { return nil }
