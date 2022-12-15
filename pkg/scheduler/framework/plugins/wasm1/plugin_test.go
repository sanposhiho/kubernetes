package wasm

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	st "k8s.io/kubernetes/pkg/scheduler/testing"
)

func Test_Filter(t *testing.T) {
	tests := []struct {
		name         string
		pod          *v1.Pod
		node         *v1.Node
		expectedCode framework.Code
	}{
		{
			name: "success: good-node",
			pod:  st.MakePod().Name("p").Obj(),

			node:         st.MakeNode().Name("good-node").Obj(),
			expectedCode: framework.Success,
		},
		{
			name: "filtered: bad-node",
			pod:  st.MakePod().Name("p").Obj(),

			node:         st.MakeNode().Name("bad-node").Obj(),
			expectedCode: framework.Unschedulable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, err := New(nil, nil)
			if err != nil {
				t.Fatalf("failed to create plugin: %v", err)
			}

			ni := framework.NewNodeInfo()
			ni.SetNode(test.node)
			s := p.(framework.FilterPlugin).Filter(context.Background(), nil, test.pod, ni)
			if s.Code() != test.expectedCode {
				t.Fatalf("unexpected code: got %v, expected %v, got reason: %v", s.Code(), test.expectedCode, s.Message())
			}
		})
	}
}
