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

package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2/ktesting"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	st "k8s.io/kubernetes/pkg/scheduler/testing"
	testingclock "k8s.io/utils/clock/testing"
)

const queueMetricMetadata = `
		# HELP scheduler_queue_incoming_pods_total [STABLE] Number of pods added to scheduling queues by event and queue type.
		# TYPE scheduler_queue_incoming_pods_total counter
	`

var (
	TestEvent    = framework.ClusterEvent{Resource: "test"}
	NodeAllEvent = framework.ClusterEvent{Resource: framework.Node, ActionType: framework.All}
	EmptyEvent   = framework.ClusterEvent{}

	lowPriority, midPriority, highPriority = int32(0), int32(100), int32(1000)
	mediumPriority                         = (lowPriority + highPriority) / 2

	highPriorityPodInfo = mustNewPodInfo(
		st.MakePod().Name("hpp").Namespace("ns1").UID("hppns1").Priority(highPriority).Obj(),
	)
	highPriNominatedPodInfo = mustNewPodInfo(
		st.MakePod().Name("hpp").Namespace("ns1").UID("hppns1").Priority(highPriority).NominatedNodeName("node1").Obj(),
	)
	medPriorityPodInfo = mustNewPodInfo(
		st.MakePod().Name("mpp").Namespace("ns2").UID("mppns2").Annotation("annot2", "val2").Priority(mediumPriority).NominatedNodeName("node1").Obj(),
	)
	unschedulablePodInfo = mustNewPodInfo(
		st.MakePod().Name("up").Namespace("ns1").UID("upns1").Annotation("annot2", "val2").Priority(lowPriority).NominatedNodeName("node1").Condition(v1.PodScheduled, v1.ConditionFalse, v1.PodReasonUnschedulable).Obj(),
	)
	nonExistentPodInfo = mustNewPodInfo(
		st.MakePod().Name("ne").Namespace("ns1").UID("nens1").Obj(),
	)
	scheduledPodInfo = mustNewPodInfo(
		st.MakePod().Name("sp").Namespace("ns1").UID("spns1").Node("foo").Obj(),
	)

	nominatorCmpOpts = []cmp.Option{
		cmp.AllowUnexported(nominator{}),
		cmpopts.IgnoreFields(nominator{}, "podLister", "lock"),
	}

	queueHintReturnQueueAfterBackoff = func(pod *v1.Pod, oldObj, newObj interface{}) framework.QueueingHint {
		return framework.QueueAfterBackoff
	}
	queueHintReturnQueueImmediately = func(pod *v1.Pod, oldObj, newObj interface{}) framework.QueueingHint {
		return framework.QueueImmediately
	}
	queueHintReturnQueueSkip = func(pod *v1.Pod, oldObj, newObj interface{}) framework.QueueingHint {
		return framework.QueueSkip
	}
)

func mustNewPodInfo(pod *v1.Pod) *framework.PodInfo {
	podInfo, err := framework.NewPodInfo(pod)
	if err != nil {
		panic(err)
	}
	return podInfo
}

func getUnschedulablePod(p *PriorityQueue, pod *v1.Pod) *v1.Pod {
	pInfo := p.unschedulablePods.get(pod)
	if pInfo != nil {
		return pInfo.Pod
	}
	return nil
}

// makeEmptyQueueingHintMapPerProfile initializes an empty QueueingHintMapPerProfile for "" profile name.
func makeEmptyQueueingHintMapPerProfile() QueueingHintMapPerProfile {
	m := make(QueueingHintMapPerProfile)
	return m
}

func BenchmarkMoveAllToActiveOrBackoffQueue(b *testing.B) {
	tests := []struct {
		name      string
		moveEvent framework.ClusterEvent
	}{
		{
			name:      "baseline",
			moveEvent: UnschedulableTimeout,
		},
		{
			name:      "worst",
			moveEvent: NodeAdd,
		},
		{
			name: "random",
			// leave "moveEvent" unspecified
		},
	}

	podTemplates := []*v1.Pod{
		highPriorityPodInfo.Pod, highPriNominatedPodInfo.Pod,
		medPriorityPodInfo.Pod, unschedulablePodInfo.Pod,
	}

	events := []framework.ClusterEvent{
		NodeAdd,
		NodeTaintChange,
		NodeAllocatableChange,
		NodeConditionChange,
		NodeLabelChange,
		PvcAdd,
		PvcUpdate,
		PvAdd,
		PvUpdate,
		StorageClassAdd,
		StorageClassUpdate,
		CSINodeAdd,
		CSINodeUpdate,
		CSIDriverAdd,
		CSIDriverUpdate,
		CSIStorageCapacityAdd,
		CSIStorageCapacityUpdate,
	}

	pluginNum := 20
	var plugins []string
	// Mimic that we have 20 plugins loaded in runtime.
	for i := 0; i < pluginNum; i++ {
		plugins = append(plugins, fmt.Sprintf("fake-plugin-%v", i))
	}

	for _, tt := range tests {
		for _, podsInUnschedulablePods := range []int{1000, 5000} {
			b.Run(fmt.Sprintf("%v-%v", tt.name, podsInUnschedulablePods), func(b *testing.B) {
				logger, _ := ktesting.NewTestContext(b)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					c := testingclock.NewFakeClock(time.Now())

					m := makeEmptyQueueingHintMapPerProfile()
					// - All plugins registered for events[0], which is NodeAdd.
					// - 1/2 of plugins registered for events[1]
					// - 1/3 of plugins registered for events[2]
					// - ...
					for j := 0; j < len(events); j++ {
						for k := 0; k < len(plugins); k++ {
							if (k+1)%(j+1) == 0 {
								found := false
								for i := range m[""] {
									if m[""][i].Event == events[j] {
										found = true
										m[""][i].QueueingHintFn = append(m[""][i].QueueingHintFn,
											&QueueingHintFunction{
												PluginName:     plugins[k],
												QueueingHintFn: queueHintReturnQueueAfterBackoff,
											},
										)
									}
								}
								if !found {
									m[""] = append(m[""], &QueueingHintWithEvent{
										Event: events[j],
										QueueingHintFn: []*QueueingHintFunction{
											{
												PluginName:     plugins[k],
												QueueingHintFn: queueHintReturnQueueAfterBackoff,
											},
										},
									})
								}
							}
						}
					}

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c), WithQueueingHintMapPerProfile(m))

					// Init pods in unschedulablePods.
					for j := 0; j < podsInUnschedulablePods; j++ {
						p := podTemplates[j%len(podTemplates)].DeepCopy()
						p.Name, p.UID = fmt.Sprintf("%v-%v", p.Name, j), types.UID(fmt.Sprintf("%v-%v", p.UID, j))
						var podInfo *framework.QueuedPodInfo
						// The ultimate goal of composing each PodInfo is to cover the path that intersects
						// (unschedulable) plugin names with the plugins that register the moveEvent,
						// here the rational is:
						// - in baseline case, don't inject unschedulable plugin names, so podMatchesEvent()
						//   never gets executed.
						// - in worst case, make both ends (of the intersection) a big number,i.e.,
						//   M intersected with N instead of M with 1 (or 1 with N)
						// - in random case, each pod failed by a random plugin, and also the moveEvent
						//   is randomized.
						if tt.name == "baseline" {
							podInfo = q.newQueuedPodInfo(p)
						} else if tt.name == "worst" {
							// Each pod failed by all plugins.
							podInfo = q.newQueuedPodInfo(p, plugins...)
						} else {
							// Random case.
							podInfo = q.newQueuedPodInfo(p, plugins[j%len(plugins)])
						}
						q.AddUnschedulableIfNotPresent(logger, podInfo, q.SchedulingCycle())
					}

					b.StartTimer()
					if tt.moveEvent.Resource != "" {
						q.MoveAllToActiveOrBackoffQueue(logger, tt.moveEvent, nil, nil, nil)
					} else {
						// Random case.
						q.MoveAllToActiveOrBackoffQueue(logger, events[i%len(events)], nil, nil, nil)
					}
				}
			})
		}
	}
}

func newDefaultQueueSort() framework.LessFunc {
	sort := &queuesort.PrioritySort{}
	return sort.Less
}
