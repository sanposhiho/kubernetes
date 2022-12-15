/*
Copyright 2019 The Kubernetes Authors.

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

package interpodaffinity

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	plugintesting "k8s.io/kubernetes/pkg/scheduler/framework/plugins/testing"
	"k8s.io/kubernetes/pkg/scheduler/internal/cache"
	st "k8s.io/kubernetes/pkg/scheduler/testing"
)

func Benchmark_Filter(b *testing.B) {
	podLabel := map[string]string{"service": "securityscan"}
	pod := st.MakePod().Labels(podLabel).Node("node1").Obj()

	labels1 := map[string]string{
		"region": "r1",
		"zone":   "z11",
	}
	podLabel2 := map[string]string{"security": "S1"}
	node1 := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: labels1}}

	tests := []struct {
		name         string
		pod          *v1.Pod
		pods         []*v1.Pod
		node         *v1.Node
		expectedCode framework.Code
	}{
		{
			name:         "satisfies with requiredDuringSchedulingIgnoredDuringExecution in PodAffinity using In operator that matches the existing pod",
			pod:          st.MakePod().Namespace(defaultNamespace).Labels(podLabel2).PodAffinityIn("service", "region", []string{"securityscan", "value2"}, st.PodAffinityWithRequiredReq).Obj(),
			pods:         []*v1.Pod{pod},
			node:         &node1,
			expectedCode: framework.Success,
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			snapshot := cache.NewSnapshot(test.pods, []*v1.Node{test.node})
			p := plugintesting.SetupPluginWithInformers(ctx, b, New, &config.InterPodAffinityArgs{}, snapshot, namespaces)
			state := framework.NewCycleState()
			_, preFilterStatus := p.(framework.PreFilterPlugin).PreFilter(context.Background(), state, test.pod)
			if !preFilterStatus.IsSuccess() {
				b.Errorf("prefilter failed with status: %v", preFilterStatus)
			}


			ni := framework.NewNodeInfo()
			ni.SetNode(test.node)
			for i := 0; i < b.N; i++ {
				s := p.(framework.FilterPlugin).Filter(context.Background(), state, test.pod, ni)
				if s.Code() != test.expectedCode {
					b.Fatalf("unexpected code: got %v, expected %v, got reason: %v", s.Code(), test.expectedCode, s.Message())
				}
			}
		})
	}
}
