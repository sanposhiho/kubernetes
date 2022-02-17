/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"k8s.io/kube-scheduler/config/v1beta3"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	configtesting "k8s.io/kubernetes/pkg/scheduler/apis/config/testing"
	"k8s.io/utils/pointer"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	tef "k8s.io/kubernetes/test/integration/framework"
	testutils "k8s.io/kubernetes/test/integration/util"
)

func main() {
	tef.EtcdMain(Run)
}

func initTest(t *testing.T, nsPrefix string, opts ...scheduler.Option) *testutils.TestContext {
	testCtx := testutils.InitTestSchedulerWithOptions(t, testutils.InitTestAPIServer(t, nsPrefix, nil), opts...)
	testutils.SyncInformerFactory(testCtx)
	go testCtx.Scheduler.Run(testCtx.Ctx)
	return testCtx
}

func Run() int {
	t := &testing.T{}

	pf := os.Getenv("PLUGIN_FILE")
	cfg := configtesting.V1beta3ToInternalWithDefaults(t, v1beta3.KubeSchedulerConfiguration{
		Profiles: []v1beta3.KubeSchedulerProfile{{
			SchedulerName: pointer.StringPtr("default-scheduler"),
		},
		},
	})
	cfg.Profiles[0].Plugins.Score.Enabled = append(cfg.Profiles[0].Plugins.Score.Enabled, config.Plugin{
		Name:   "nodenumber",
		Weight: 10,
	})

	testCtx := initTest(t, "hoge",
		scheduler.WithPodInitialBackoffSeconds(0),
		scheduler.WithPodMaxBackoffSeconds(0),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithCustomPluginFiles(map[string]string{
			"nodenumber": pf,
		}),
	)

	klog.Info("HOGE")
	testutils.SyncInformerFactory(testCtx)

	defer testutils.CleanupTest(t, testCtx)

	cs, ns, ctx := testCtx.ClientSet, testCtx.NS.Name, testCtx.Ctx
	cpu, _ := resource.ParseQuantity("4")
	mem, _ := resource.ParseQuantity("32Gi")
	podnum, _ := resource.ParseQuantity("10")

	// create node0 ~ node9, but all nodes
	for i := 0; i < 9; i++ {
		suffix := strconv.Itoa(i)
		_, err := cs.CoreV1().Nodes().Create(ctx, &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node" + suffix,
			},
			Status: v1.NodeStatus{Capacity: v1.ResourceList{
				v1.ResourceCPU:    cpu,
				v1.ResourceMemory: mem,
				v1.ResourcePods:   podnum,
			}},
			Spec: v1.NodeSpec{},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("create node: %w", err)
		}
	}
	klog.Info("HOGE4")

	// pod1 should be bound to node1 because of nodenumber plugin.
	if _, err := cs.CoreV1().Pods(ns).Create(ctx, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container1",
					Image: "k8s.gcr.io/pause:3.5",
				},
			},
		},
	}, metav1.CreateOptions{}); err != nil {
		klog.Errorf("Failed to create Pod %v", err)
		return 3
	}

	time.Sleep(10 * time.Second)

	klog.Info("HOGE3")
	p, err := cs.CoreV1().Pods(ns).Get(ctx, "pod1", metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return 4
	}
	klog.Info("HOGE0")

	if p.Spec.NodeName != "node1" {
		klog.Errorf("non expected nodename: %v", p)
		return 2
	}

	klog.Info("HOGE2")

	panic("end")
	return 0
}

// nextPodOrDie returns the next Pod in the scheduler queue.
// The operation needs to be completed within 5 seconds; otherwise the test gets aborted.
func nextPodOrDie(t *testing.T, testCtx *testutils.TestContext) *framework.QueuedPodInfo {
	t.Helper()

	var podInfo *framework.QueuedPodInfo
	// NextPod() is a blocking operation. Wrap it in timeout() to avoid relying on
	// default go testing timeout (10m) to abort.
	if err := timeout(testCtx.Ctx, time.Second*5, func() {
		podInfo = testCtx.Scheduler.NextPod()
	}); err != nil {
		panic("timeout: %v" + err.Error())
	}
	return podInfo
}

// timeout returns a timeout error if the given `f` function doesn't
// complete within `d` duration; otherwise it returns nil.
func timeout(ctx context.Context, d time.Duration, f func()) error {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	done := make(chan struct{})
	go func() {
		f()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
