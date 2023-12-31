/*
Copyright 2021 The Kubernetes Authors.

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
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"
	apiservertesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/scheduler"
	configtesting "k8s.io/kubernetes/pkg/scheduler/apis/config/testing"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/names"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	st "k8s.io/kubernetes/pkg/scheduler/testing"
	testfwk "k8s.io/kubernetes/test/integration/framework"
	testutils "k8s.io/kubernetes/test/integration/util"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"k8s.io/utils/pointer"
)

func TestSchedulingGates(t *testing.T) {
	tests := []struct {
		name                  string
		pods                  []*v1.Pod
		featureEnabled        bool
		want                  []string
		rmPodsSchedulingGates []int
		wantPostGatesRemoval  []string
	}{
		{
			name: "feature disabled, regular pods",
			pods: []*v1.Pod{
				st.MakePod().Name("p1").Container("pause").Obj(),
				st.MakePod().Name("p2").Container("pause").Obj(),
			},
			featureEnabled: false,
			want:           []string{"p1", "p2"},
		},
		{
			name: "feature enabled, regular pods",
			pods: []*v1.Pod{
				st.MakePod().Name("p1").Container("pause").Obj(),
				st.MakePod().Name("p2").Container("pause").Obj(),
			},
			featureEnabled: true,
			want:           []string{"p1", "p2"},
		},
		{
			name: "feature disabled, one pod carrying scheduling gates",
			pods: []*v1.Pod{
				st.MakePod().Name("p1").SchedulingGates([]string{"foo"}).Container("pause").Obj(),
				st.MakePod().Name("p2").Container("pause").Obj(),
			},
			featureEnabled: false,
			want:           []string{"p1", "p2"},
		},
		{
			name: "feature enabled, one pod carrying scheduling gates",
			pods: []*v1.Pod{
				st.MakePod().Name("p1").SchedulingGates([]string{"foo"}).Container("pause").Obj(),
				st.MakePod().Name("p2").Container("pause").Obj(),
			},
			featureEnabled: true,
			want:           []string{"p2"},
		},
		{
			name: "feature enabled, two pod carrying scheduling gates, and remove gates of one pod",
			pods: []*v1.Pod{
				st.MakePod().Name("p1").SchedulingGates([]string{"foo"}).Container("pause").Obj(),
				st.MakePod().Name("p2").SchedulingGates([]string{"bar"}).Container("pause").Obj(),
				st.MakePod().Name("p3").Container("pause").Obj(),
			},
			featureEnabled:        true,
			want:                  []string{"p3"},
			rmPodsSchedulingGates: []int{1}, // remove gates of 'p2'
			wantPostGatesRemoval:  []string{"p2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.PodSchedulingReadiness, tt.featureEnabled)()

			// Use zero backoff seconds to bypass backoffQ.
			// It's intended to not start the scheduler's queue, and hence to
			// not start any flushing logic. We will pop and schedule the Pods manually later.
			testCtx := testutils.InitTestSchedulerWithOptions(
				t,
				testutils.InitTestAPIServer(t, "pod-scheduling-gates", nil),
				0,
				scheduler.WithPodInitialBackoffSeconds(0),
				scheduler.WithPodMaxBackoffSeconds(0),
			)
			testutils.SyncSchedulerInformerFactory(testCtx)

			cs, ns, ctx := testCtx.ClientSet, testCtx.NS.Name, testCtx.Ctx
			for _, p := range tt.pods {
				p.Namespace = ns
				if _, err := cs.CoreV1().Pods(ns).Create(ctx, p, metav1.CreateOptions{}); err != nil {
					t.Fatalf("Failed to create Pod %q: %v", p.Name, err)
				}
			}

			// Wait for the pods to be present in the scheduling queue.
			if err := wait.PollUntilContextTimeout(ctx, time.Millisecond*200, wait.ForeverTestTimeout, false, func(ctx context.Context) (bool, error) {
				pendingPods, _ := testCtx.Scheduler.SchedulingQueue.PendingPods()
				return len(pendingPods) == len(tt.pods), nil
			}); err != nil {
				t.Fatal(err)
			}

			// Pop the expected pods out. They should be de-queueable.
			for _, wantPod := range tt.want {
				podInfo := testutils.NextPodOrDie(t, testCtx)
				if got := podInfo.Pod.Name; got != wantPod {
					t.Errorf("Want %v to be popped out, but got %v", wantPod, got)
				}
			}

			if len(tt.rmPodsSchedulingGates) == 0 {
				return
			}
			// Remove scheduling gates from the pod spec.
			for _, idx := range tt.rmPodsSchedulingGates {
				patch := `{"spec": {"schedulingGates": null}}`
				podName := tt.pods[idx].Name
				if _, err := cs.CoreV1().Pods(ns).Patch(ctx, podName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{}); err != nil {
					t.Fatalf("Failed to patch pod %v: %v", podName, err)
				}
			}
			// Pop the expected pods out. They should be de-queueable.
			for _, wantPod := range tt.wantPostGatesRemoval {
				podInfo := testutils.NextPodOrDie(t, testCtx)
				if got := podInfo.Pod.Name; got != wantPod {
					t.Errorf("Want %v to be popped out, but got %v", wantPod, got)
				}
			}
		})
	}
}

// TestCoreResourceEnqueue verify Pods failed by in-tree default plugins can be
// moved properly upon their registered events.
func TestCoreResourceEnqueue(t *testing.T) {
	// Use zero backoff seconds to bypass backoffQ.
	// It's intended to not start the scheduler's queue, and hence to
	// not start any flushing logic. We will pop and schedule the Pods manually later.
	testCtx := testutils.InitTestSchedulerWithOptions(
		t,
		testutils.InitTestAPIServer(t, "core-res-enqueue", nil),
		0,
		scheduler.WithPodInitialBackoffSeconds(0),
		scheduler.WithPodMaxBackoffSeconds(0),
	)
	testutils.SyncSchedulerInformerFactory(testCtx)

	defer testCtx.Scheduler.SchedulingQueue.Close()

	cs, ns, ctx := testCtx.ClientSet, testCtx.NS.Name, testCtx.Ctx
	// Create one Node with a taint.
	node := st.MakeNode().Label("node", "fake-node").Name("fake-node").Capacity(map[v1.ResourceName]string{v1.ResourceCPU: "2"}).Obj()
	node.Spec.Taints = []v1.Taint{{Key: v1.TaintNodeNotReady, Effect: v1.TaintEffectNoSchedule}}
	if _, err := cs.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Node %q: %v", node.Name, err)
	}

	// Create two Pods that are both unschedulable.
	// - Pod1 and Pod2 have the same hostport that is currently available, but doesn't have the required toleration. (unschedPlugin: TaintToleration)
	//   They cannot be scheduled on the same node due to PodAffinity.
	// - Pod3 requests a large amount of CPU resource that the node cannot fit. (unschedPlugin: TaintToleration)
	//   Note: Pod2 will fail the tainttoleration plugin b/c that's ordered prior to noderesources.
	// - Pod4 has the required toleration, but requests a large amount of CPU. (unschedPlugin: NodeResourcesFit)
	pod1 := st.MakePod().Label("anti", "anti").Namespace(ns).Name("pod1").Priority(1000).PodAntiAffinityExists("anti", "node", st.PodAntiAffinityWithRequiredReq).Container("image").Obj()
	pod2 := st.MakePod().Label("anti", "anti").Namespace(ns).Name("pod2").Priority(1).PodAntiAffinityExists("anti", "node", st.PodAntiAffinityWithRequiredReq).Container("image").Obj()
	pod3 := st.MakePod().Namespace(ns).Name("pod3").Req(map[v1.ResourceName]string{v1.ResourceCPU: "4"}).Obj()
	pod4 := st.MakePod().Namespace(ns).Name("pod4").Toleration(v1.TaintNodeNotReady).Req(map[v1.ResourceName]string{v1.ResourceCPU: "4"}).Container("image").Obj()
	for _, pod := range []*v1.Pod{pod1, pod2, pod3, pod4} {
		if _, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
			t.Fatalf("Failed to create Pod %q: %v", pod.Name, err)
		}
	}

	// Wait for the three pods to be present in the scheduling queue.
	if err := wait.PollUntilContextTimeout(ctx, time.Millisecond*200, wait.ForeverTestTimeout, false, func(ctx context.Context) (bool, error) {
		pendingPods, _ := testCtx.Scheduler.SchedulingQueue.PendingPods()
		return len(pendingPods) == 4, nil
	}); err != nil {
		t.Fatal(err)
	}

	t.Log("Confirmed Pods in the scheduling queue, starting to schedule them")

	// Pop all pods out. They should be unschedulable.
	scheduleDone := sets.New[string]()
	for i := 0; i < 4; i++ {
		scheduled := scheduleNextPod(t, testCtx, false /* all Pods are supposed to fail at scheduling */)
		if scheduleDone.Has(scheduled) {
			t.Fatalf("Pod %v is scheduled twice", scheduled)
		}

		scheduleDone.Insert(scheduled)
	}

	t.Log("finished initial schedulings for all Pods, will remove taint from fake-node")

	// Trigger a NodeTaintChange event.
	// We expect this event to trigger moving Pod1/Pod2 from unschedulablePods to activeQ.
	// Pod3 and Pod4 are not expected to be popped out.
	// - Although the failure reason has been lifted, Pod3 still won't be moved to active due to
	//   the node event's preCheckForNode() because Pod4 doesn't have enough CPU.
	// - Regarding Pod4, the NodeTaintChange event is filtered out preCheckForNode() because the node doesn't have enough CPU.
	node.Spec.Taints = nil
	if _, err := cs.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("Failed to remove taints off the node: %v", err)
	}

	t.Log("removed taint from fake-node, will confirm Pod1 are moved to activeQ")

	// Try to schedule next Pod, which should be Pod1 because Pod1 has a higher priority than Pod2.
	// Pod1 should be scheduled successfully.
	scheduled := scheduleNextPod(t, testCtx, true)
	if scheduled != "pod1" {
		t.Fatalf("Expect Pod1 to be scheduled, but got %v", scheduled)
	}

	t.Log("confirmed Pod1 is scheduled, will confirm Pod2 is moved to activeQ")

	// Try to schedule next Pod, which should be Pod2.
	// But, Pod2 should be unschedulable again because Pod1, which has the Pod anti-affinity to Pod2, is already scheduled.
	scheduled = scheduleNextPod(t, testCtx, false)
	if scheduled != "pod2" {
		t.Fatalf("Expect Pod2 to be scheduled, but got %v", scheduled)
	}

	t.Log("confirmed Pod2 was moved to activeQ but the scheduling was failed, will confirm Pod3 and Pod4 are NOT moved to activeQ")

	// Ensure we can pop no more Pods out at the moment.
	if len(testCtx.Scheduler.SchedulingQueue.PodsInActiveQ()) != 0 {
		t.Fatalf("Expect no Pods in activeQ, but %v Pods are in activeQ", len(testCtx.Scheduler.SchedulingQueue.PodsInActiveQ()))
	}

	// Current situation:
	// - Pod1 is scheduled.
	// - Pod2 is unschedulable and has unschedPlugin: NodePort.
	// - Pod3 is unschedulable and had unschedPlugin: TaintToleration.
	// - Pod4 is unschedulable and had unschedPlugin: NodeResourcesFit.

	t.Log("Create fake-node2 to make all remaining Pods schedulable")

	// Create another Node with a taint.
	// This Pod can accommodate all Pods once the taint is removed.
	node2 := st.MakeNode().Label("node", "fake-node2").Name("fake-node2").Capacity(map[v1.ResourceName]string{v1.ResourceCPU: "8"}).Obj()
	node2.Spec.Taints = []v1.Taint{{Key: v1.TaintNodeNotReady, Effect: v1.TaintEffectNoSchedule}}
	if _, err := cs.CoreV1().Nodes().Create(ctx, node2, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Node %q: %v", node2.Name, err)
	}

	t.Log("fake-node2 is created, will confirm Pod4 is moved to activeQ")

	// Try to schedule next Pod, which should be Pod4 because Pod4 has a toleration.
	// As mentioned above, this scheduling should be successful.
	scheduled = scheduleNextPod(t, testCtx, true)
	if scheduled != "pod4" {
		t.Fatalf("Expect Pod2 to be scheduled, but got %v", scheduled)
	}

	t.Log("confirmed Pod4 is scheduled, will confirm Pod2 and Pod3 are NOT moved to activeQ")

	// Ensure we can pop no more Pods out at the moment because Pod2 and Pod3 don't have toleration.
	if len(testCtx.Scheduler.SchedulingQueue.PodsInActiveQ()) != 0 {
		t.Fatalf("Expect no Pods in activeQ, but %v Pods are in activeQ", len(testCtx.Scheduler.SchedulingQueue.PodsInActiveQ()))
	}

	t.Log("will remove taint from fake-node2")

	// Trigger a NodeTaintChange event.
	// Then Pod2 and Pod3 should be moved to activeQ.
	node2.Spec.Taints = nil
	if _, err := cs.CoreV1().Nodes().Update(ctx, node2, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("Failed to remove taints off the node2: %v", err)
	}

	t.Log("the taint is removed from fake-node2, will confirm Pod2 and Pod3 are moved to activeQ and scheduled successfully")

	// Pop all pods out. They should be schedulable.
	scheduleDone = sets.New[string]()
	for i := 0; i < 2; i++ {
		scheduled := scheduleNextPod(t, testCtx, true /* all Pods are supposed to succeed in scheduling */)
		if scheduleDone.Has(scheduled) {
			t.Fatalf("Pod %v is scheduled twice", scheduled)
		}

		scheduleDone.Insert(scheduled)
	}

	if scheduleDone.Difference(sets.New("pod2", "pod3")).Len() != 0 {
		t.Fatalf("Expect Pod2 and Pod3 to be scheduled, but got %v", scheduleDone)
	}
}

// scheduleNextPod pops the next Pod from the scheduling queue and schedule it manually.
// It returns the name of the Pod that was popped.
func scheduleNextPod(t *testing.T, testCtx *testutils.TestContext, scheduleSuccess bool) string {
	podInfo := testutils.NextPodOrDie(t, testCtx)
	t.Logf("Pop Pod %v out, will schedule it", podInfo.Pod.Name)
	fwk, ok := testCtx.Scheduler.Profiles[podInfo.Pod.Spec.SchedulerName]
	if !ok {
		t.Fatalf("Cannot find the profile for Pod %v", podInfo.Pod.Name)
	}
	r, fitError := testCtx.Scheduler.SchedulePod(testCtx.Ctx, fwk, framework.NewCycleState(), podInfo.Pod)
	if scheduleSuccess {
		if fitError != nil {
			t.Fatalf("Expect Pod %v to be scheduled", podInfo.Pod.Name)
		}

		// Assume this Pod to simulate this Pod goes to the binding cycle.
		pod := podInfo.Pod.DeepCopy()
		pod.Spec.NodeName = r.SuggestedHost
		if err := testCtx.Scheduler.Cache.AssumePod(klog.FromContext(testCtx.Ctx), pod); err != nil {
			t.Fatalf("Failed to assume Pod %v: %v", pod.Name, err)
		}

		return podInfo.Pod.Name
	}

	if fitError == nil {
		t.Fatalf("Expect Pod %v to fail at scheduling, but scheduled to %s", podInfo.Pod.Name, r.SuggestedHost)
	}
	testCtx.Scheduler.FailureHandler(testCtx.Ctx, fwk, podInfo, framework.NewStatus(framework.Unschedulable).WithError(fitError), nil, time.Now())
	return podInfo.Pod.Name
}

var _ framework.FilterPlugin = &fakeCRPlugin{}
var _ framework.EnqueueExtensions = &fakeCRPlugin{}

type fakeCRPlugin struct{}

func (f *fakeCRPlugin) Name() string {
	return "fakeCRPlugin"
}

func (f *fakeCRPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return framework.NewStatus(framework.Unschedulable, "always fail")
}

// EventsToRegister returns the possible events that may make a Pod
// failed by this plugin schedulable.
func (f *fakeCRPlugin) EventsToRegister() []framework.ClusterEventWithHint {
	return []framework.ClusterEventWithHint{
		{Event: framework.ClusterEvent{Resource: "foos.v1.example.com", ActionType: framework.All}},
	}
}

// TestCustomResourceEnqueue constructs a fake plugin that registers custom resources
// to verify Pods failed by this plugin can be moved properly upon CR events.
func TestCustomResourceEnqueue(t *testing.T) {
	// Start API Server with apiextensions supported.
	server := apiservertesting.StartTestServerOrDie(
		t, apiservertesting.NewDefaultTestServerOptions(),
		[]string{"--disable-admission-plugins=ServiceAccount,TaintNodesByCondition", "--runtime-config=api/all=true"},
		testfwk.SharedEtcd(),
	)
	testCtx := &testutils.TestContext{}
	ctx, cancel := context.WithCancel(context.Background())
	testCtx.Ctx = ctx
	testCtx.CloseFn = func() {
		cancel()
		server.TearDownFn()
	}

	apiExtensionClient := apiextensionsclient.NewForConfigOrDie(server.ClientConfig)
	dynamicClient := dynamic.NewForConfigOrDie(server.ClientConfig)

	// Create a Foo CRD.
	fooCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foos.example.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural: "foos",
				Kind:   "Foo",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"field": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}
	var err error
	fooCRD, err = apiExtensionClient.ApiextensionsV1().CustomResourceDefinitions().Create(testCtx.Ctx, fooCRD, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	registry := frameworkruntime.Registry{
		"fakeCRPlugin": func(_ context.Context, _ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
			return &fakeCRPlugin{}, nil
		},
	}
	cfg := configtesting.V1ToInternalWithDefaults(t, configv1.KubeSchedulerConfiguration{
		Profiles: []configv1.KubeSchedulerProfile{{
			SchedulerName: pointer.String(v1.DefaultSchedulerName),
			Plugins: &configv1.Plugins{
				Filter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "fakeCRPlugin"},
					},
				},
			},
		}}})

	testCtx.KubeConfig = server.ClientConfig
	testCtx.ClientSet = kubernetes.NewForConfigOrDie(server.ClientConfig)
	testCtx.NS, err = testCtx.ClientSet.CoreV1().Namespaces().Create(testCtx.Ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("cr-enqueue-%v", string(uuid.NewUUID()))}}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		t.Fatalf("Failed to integration test ns: %v", err)
	}

	// Use zero backoff seconds to bypass backoffQ.
	// It's intended to not start the scheduler's queue, and hence to
	// not start any flushing logic. We will pop and schedule the Pods manually later.
	testCtx = testutils.InitTestSchedulerWithOptions(
		t,
		testCtx,
		0,
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithFrameworkOutOfTreeRegistry(registry),
		scheduler.WithPodInitialBackoffSeconds(0),
		scheduler.WithPodMaxBackoffSeconds(0),
	)
	testutils.SyncSchedulerInformerFactory(testCtx)

	defer testutils.CleanupTest(t, testCtx)

	cs, ns, ctx := testCtx.ClientSet, testCtx.NS.Name, testCtx.Ctx
	logger := klog.FromContext(ctx)
	// Create one Node.
	node := st.MakeNode().Name("fake-node").Obj()
	if _, err := cs.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Node %q: %v", node.Name, err)
	}

	// Create a testing Pod.
	pause := imageutils.GetPauseImageName()
	pod := st.MakePod().Namespace(ns).Name("fake-pod").Container(pause).Obj()
	if _, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pod %q: %v", pod.Name, err)
	}

	// Wait for the testing Pod to be present in the scheduling queue.
	if err := wait.PollUntilContextTimeout(ctx, time.Millisecond*200, wait.ForeverTestTimeout, false, func(ctx context.Context) (bool, error) {
		pendingPods, _ := testCtx.Scheduler.SchedulingQueue.PendingPods()
		return len(pendingPods) == 1, nil
	}); err != nil {
		t.Fatal(err)
	}

	// Pop fake-pod out. It should be unschedulable.
	podInfo := testutils.NextPodOrDie(t, testCtx)
	fwk, ok := testCtx.Scheduler.Profiles[podInfo.Pod.Spec.SchedulerName]
	if !ok {
		t.Fatalf("Cannot find the profile for Pod %v", podInfo.Pod.Name)
	}
	// Schedule the Pod manually.
	_, fitError := testCtx.Scheduler.SchedulePod(ctx, fwk, framework.NewCycleState(), podInfo.Pod)
	// The fitError is expected to be non-nil as it failed the fakeCRPlugin plugin.
	if fitError == nil {
		t.Fatalf("Expect Pod %v to fail at scheduling.", podInfo.Pod.Name)
	}
	testCtx.Scheduler.FailureHandler(ctx, fwk, podInfo, framework.NewStatus(framework.Unschedulable).WithError(fitError), nil, time.Now())

	// Scheduling cycle is incremented from 0 to 1 after NextPod() is called, so
	// pass a number larger than 1 to move Pod to unschedulablePods.
	testCtx.Scheduler.SchedulingQueue.AddUnschedulableIfNotPresent(logger, podInfo, 10)

	// Trigger a Custom Resource event.
	// We expect this event to trigger moving the test Pod from unschedulablePods to activeQ.
	crdGVR := schema.GroupVersionResource{Group: fooCRD.Spec.Group, Version: fooCRD.Spec.Versions[0].Name, Resource: "foos"}
	crClient := dynamicClient.Resource(crdGVR).Namespace(ns)
	if _, err := crClient.Create(ctx, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       "Foo",
			"metadata":   map[string]interface{}{"name": "foo1"},
		},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Unable to create cr: %v", err)
	}

	// Now we should be able to pop the Pod from activeQ again.
	podInfo = testutils.NextPodOrDie(t, testCtx)
	if podInfo.Attempts != 2 {
		t.Errorf("Expected the Pod to be attempted 2 times, but got %v", podInfo.Attempts)
	}
}

// TestRequeueByBindFailure verify Pods failed by bind plugin are
// put back to the queue regardless of whether event happens or not.
func TestRequeueByBindFailure(t *testing.T) {
	fakeBind := &firstFailBindPlugin{}
	registry := frameworkruntime.Registry{
		"firstFailBindPlugin": func(ctx context.Context, o runtime.Object, fh framework.Handle) (framework.Plugin, error) {
			binder, err := defaultbinder.New(ctx, nil, fh)
			if err != nil {
				return nil, err
			}

			fakeBind.defaultBinderPlugin = binder.(framework.BindPlugin)
			return fakeBind, nil
		},
	}

	cfg := configtesting.V1ToInternalWithDefaults(t, configv1.KubeSchedulerConfiguration{
		Profiles: []configv1.KubeSchedulerProfile{{
			SchedulerName: pointer.String(v1.DefaultSchedulerName),
			Plugins: &configv1.Plugins{
				MultiPoint: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "firstFailBindPlugin"},
					},
					Disabled: []configv1.Plugin{
						{Name: names.DefaultBinder},
					},
				},
			},
		}}})

	// Use zero backoff seconds to bypass backoffQ.
	testCtx := testutils.InitTestSchedulerWithOptions(
		t,
		testutils.InitTestAPIServer(t, "core-res-enqueue", nil),
		0,
		scheduler.WithPodInitialBackoffSeconds(0),
		scheduler.WithPodMaxBackoffSeconds(0),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithFrameworkOutOfTreeRegistry(registry),
	)
	testutils.SyncSchedulerInformerFactory(testCtx)

	go testCtx.Scheduler.Run(testCtx.Ctx)

	cs, ns, ctx := testCtx.ClientSet, testCtx.NS.Name, testCtx.Ctx
	node := st.MakeNode().Name("fake-node").Obj()
	if _, err := cs.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Node %q: %v", node.Name, err)
	}
	// create a pod.
	pod := st.MakePod().Namespace(ns).Name("pod-1").Container(imageutils.GetPauseImageName()).Obj()
	if _, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pod %q: %v", pod.Name, err)
	}

	// 1. first binding try should fail.
	// 2. The pod should be enqueued to activeQ/backoffQ without any event.
	// 3. The pod should be scheduled in the second binding try.
	// Here, waiting until (3).
	err := wait.PollUntilContextTimeout(ctx, 200*time.Millisecond, wait.ForeverTestTimeout, false, testutils.PodScheduled(cs, ns, pod.Name))
	if err != nil {
		t.Fatalf("Expect pod-1 to be scheduled by the bind plugin: %v", err)
	}

	// Make sure the first binding trial was failed, and this pod is scheduled at the second trial.
	if fakeBind.counter != 1 {
		t.Fatalf("Expect pod-1 to be scheduled by the bind plugin in the second binding try: %v", err)
	}
}

// firstFailBindPlugin rejects the Pod in the first Bind call.
type firstFailBindPlugin struct {
	counter             int
	defaultBinderPlugin framework.BindPlugin
}

func (*firstFailBindPlugin) Name() string {
	return "firstFailBindPlugin"
}

func (p *firstFailBindPlugin) Bind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status {
	if p.counter == 0 {
		// fail in the first Bind call.
		p.counter++
		return framework.NewStatus(framework.Error, "firstFailBindPlugin rejects the Pod")
	}

	return p.defaultBinderPlugin.Bind(ctx, state, pod, nodename)
}

// TestRequeueByPermitRejection verify Pods failed by permit plugins in the binding cycle are
// put back to the queue, according to the correct scheduling cycle number.
func TestRequeueByPermitRejection(t *testing.T) {
	defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.SchedulerQueueingHints, true)()
	queueingHintCalledCounter := 0
	fakePermit := &fakePermitPlugin{}
	registry := frameworkruntime.Registry{
		fakePermitPluginName: func(ctx context.Context, o runtime.Object, fh framework.Handle) (framework.Plugin, error) {
			fakePermit = &fakePermitPlugin{
				frameworkHandler: fh,
				schedulingHint: func(logger klog.Logger, pod *v1.Pod, oldObj, newObj interface{}) (framework.QueueingHint, error) {
					queueingHintCalledCounter++
					return framework.Queue, nil
				},
			}
			return fakePermit, nil
		},
	}
	cfg := configtesting.V1ToInternalWithDefaults(t, configv1.KubeSchedulerConfiguration{
		Profiles: []configv1.KubeSchedulerProfile{{
			SchedulerName: pointer.String(v1.DefaultSchedulerName),
			Plugins: &configv1.Plugins{
				MultiPoint: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: fakePermitPluginName},
					},
				},
			},
		}}})

	// Use zero backoff seconds to bypass backoffQ.
	testCtx := testutils.InitTestSchedulerWithOptions(
		t,
		testutils.InitTestAPIServer(t, "core-res-enqueue", nil),
		0,
		scheduler.WithPodInitialBackoffSeconds(0),
		scheduler.WithPodMaxBackoffSeconds(0),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithFrameworkOutOfTreeRegistry(registry),
	)
	testutils.SyncSchedulerInformerFactory(testCtx)

	go testCtx.Scheduler.Run(testCtx.Ctx)

	cs, ns, ctx := testCtx.ClientSet, testCtx.NS.Name, testCtx.Ctx
	node := st.MakeNode().Name("fake-node").Obj()
	if _, err := cs.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Node %q: %v", node.Name, err)
	}
	// create a pod.
	pod := st.MakePod().Namespace(ns).Name("pod-1").Container(imageutils.GetPauseImageName()).Obj()
	if _, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pod %q: %v", pod.Name, err)
	}

	// update node label. (causes the NodeUpdate event)
	node.Labels = map[string]string{"updated": ""}
	if _, err := cs.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("Failed to add labels to the node: %v", err)
	}

	// create a pod to increment the scheduling cycle number in the scheduling queue.
	// We can make sure NodeUpdate event, that has happened in the previous scheduling cycle, makes Pod to be enqueued to activeQ via the scheduling queue.
	pod = st.MakePod().Namespace(ns).Name("pod-2").Container(imageutils.GetPauseImageName()).Obj()
	if _, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pod %q: %v", pod.Name, err)
	}

	// reject pod-1 to simulate the failure in Permit plugins.
	// This pod-1 should be enqueued to activeQ because the NodeUpdate event has happened.
	fakePermit.frameworkHandler.IterateOverWaitingPods(func(wp framework.WaitingPod) {
		if wp.GetPod().Name == "pod-1" {
			wp.Reject(fakePermitPluginName, "fakePermitPlugin rejects the Pod")
			return
		}
	})

	// Wait for pod-2 to be scheduled.
	err := wait.PollUntilContextTimeout(ctx, 200*time.Millisecond, wait.ForeverTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		fakePermit.frameworkHandler.IterateOverWaitingPods(func(wp framework.WaitingPod) {
			if wp.GetPod().Name == "pod-2" {
				wp.Allow(fakePermitPluginName)
			}
		})

		return testutils.PodScheduled(cs, ns, "pod-2")(ctx)
	})
	if err != nil {
		t.Fatalf("Expect pod-2 to be scheduled")
	}

	err = wait.PollUntilContextTimeout(ctx, 200*time.Millisecond, wait.ForeverTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		pod1Found := false
		fakePermit.frameworkHandler.IterateOverWaitingPods(func(wp framework.WaitingPod) {
			if wp.GetPod().Name == "pod-1" {
				pod1Found = true
				wp.Allow(fakePermitPluginName)
			}
		})
		return pod1Found, nil
	})
	if err != nil {
		t.Fatal("Expect pod-1 to be scheduled again")
	}

	if queueingHintCalledCounter != 1 {
		t.Fatalf("Expected the scheduling hint to be called 1 time, but %v", queueingHintCalledCounter)
	}
}

type fakePermitPlugin struct {
	frameworkHandler framework.Handle
	schedulingHint   framework.QueueingHintFn
}

const fakePermitPluginName = "fakePermitPlugin"

func (p *fakePermitPlugin) Name() string {
	return fakePermitPluginName
}

func (p *fakePermitPlugin) Permit(ctx context.Context, state *framework.CycleState, _ *v1.Pod, _ string) (*framework.Status, time.Duration) {
	return framework.NewStatus(framework.Wait), wait.ForeverTestTimeout
}

func (p *fakePermitPlugin) EventsToRegister() []framework.ClusterEventWithHint {
	return []framework.ClusterEventWithHint{
		{Event: framework.ClusterEvent{Resource: framework.Node, ActionType: framework.UpdateNodeLabel}, QueueingHintFn: p.schedulingHint},
	}
}
