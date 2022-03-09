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

package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/features"
	st "k8s.io/kubernetes/pkg/scheduler/testing"
	testutils "k8s.io/kubernetes/test/integration/util"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"k8s.io/utils/pointer"
)

// This file tests the scheduler predicates functionality.

const pollInterval = 100 * time.Millisecond

// TestInterPodAffinity verifies that scheduler's inter pod affinity and
// anti-affinity predicate functions works correctly.

// TestInterPodAffinityWithNamespaceSelector verifies that inter pod affinity with NamespaceSelector works as expected.

// TestEvenPodsSpreadPredicate verifies that EvenPodsSpread predicate functions well.
func TestEvenPodsSpreadPredicate(t *testing.T) {
	testCtx := initTest(t, "eps-predicate")
	cs := testCtx.ClientSet
	ns := testCtx.NS.Name
	defer testutils.CleanupTest(t, testCtx)

	for i := 0; i < 4; i++ {
		// Create nodes with labels "zone: zone-{0,1}" and "node: <node name>" to each node.
		nodeName := fmt.Sprintf("node-%d", i)
		zone := fmt.Sprintf("zone-%d", i/2)
		_, err := createNode(cs, st.MakeNode().Name(nodeName).Label("node", nodeName).Label("zone", zone).Obj())
		if err != nil {
			t.Fatalf("Cannot create node: %v", err)
		}
	}

	pause := imageutils.GetPauseImageName()
	tests := []struct {
		name                        string
		incomingPod                 *v1.Pod
		existingPods                []*v1.Pod
		fits                        bool
		enableMinDomainsFeatureGate bool
		candidateNodes              []string // nodes expected to schedule onto
	}{
		{
			name: "pods spread across nodes as 2/2/1, maxSkew is 2, and the number of domains < minDomains, then the third node fits",
			incomingPod: st.MakePod().Namespace(ns).Name("p").Label("foo", "").Container(pause).
				NodeAffinityNotIn("node", []string{"node-0"}). // mock a 3-node cluster
				SpreadConstraint(2, "node", hardSpread, st.MakeLabelSelector().Exists("foo").Obj(), pointer.Int32(4)).
				Obj(),
			existingPods: []*v1.Pod{
				st.MakePod().ZeroTerminationGracePeriod().Namespace(ns).Name("p1a").Node("node-1").Label("foo", "").Container(pause).Obj(),
				st.MakePod().ZeroTerminationGracePeriod().Namespace(ns).Name("p1b").Node("node-1").Label("foo", "").Container(pause).Obj(),
				st.MakePod().ZeroTerminationGracePeriod().Namespace(ns).Name("p2a").Node("node-2").Label("foo", "").Container(pause).Obj(),
				st.MakePod().ZeroTerminationGracePeriod().Namespace(ns).Name("p2b").Node("node-2").Label("foo", "").Container(pause).Obj(),
				st.MakePod().ZeroTerminationGracePeriod().Namespace(ns).Name("p3").Node("node-3").Label("foo", "").Container(pause).Obj(),
			},
			fits:                        true,
			enableMinDomainsFeatureGate: true,
			candidateNodes:              []string{"node-3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allPods := append(tt.existingPods, tt.incomingPod)
			defer testutils.CleanupPods(cs, t, allPods)
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.MinDomainsInPodTopologySpread, tt.enableMinDomainsFeatureGate)()

			for _, pod := range tt.existingPods {
				createdPod, err := cs.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Error while creating pod during test: %v", err)
				}
				err = wait.Poll(pollInterval, wait.ForeverTestTimeout, testutils.PodScheduled(cs, createdPod.Namespace, createdPod.Name))
				if err != nil {
					t.Errorf("Error while waiting for pod during test: %v", err)
				}
			}
			testPod, err := cs.CoreV1().Pods(tt.incomingPod.Namespace).Create(context.TODO(), tt.incomingPod, metav1.CreateOptions{})
			if err != nil && !apierrors.IsInvalid(err) {
				t.Fatalf("Error while creating pod during test: %v", err)
			}

			if tt.fits {
				err = wait.Poll(pollInterval, wait.ForeverTestTimeout, podScheduledIn(cs, testPod.Namespace, testPod.Name, tt.candidateNodes))
			} else {
				err = wait.Poll(pollInterval, wait.ForeverTestTimeout, podUnschedulable(cs, testPod.Namespace, testPod.Name))
			}
			if err != nil {
				t.Errorf("Test Failed: %v", err)
			}
		})
	}
}

var (
	hardSpread = v1.DoNotSchedule
	softSpread = v1.ScheduleAnyway
)
