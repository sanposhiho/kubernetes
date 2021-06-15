package node

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type Service struct {
	Client clientset.Interface
}

func NewNodeService(client clientset.Interface) *Service {
	return &Service{
		Client: client,
	}
}

func (s *Service) Create(ctx context.Context) (*v1.Node, error) {
	baseNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sample-node-",
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("32Gi"),
			},
			Phase: v1.NodeRunning,
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}

	node, err := s.Client.CoreV1().Nodes().Create(ctx, baseNode, metav1.CreateOptions{})
	if err != nil {
		return nil, xerrors.Errorf("create node: %w", err)
	}

	return node, nil
}

func (s *Service) List(ctx context.Context) (*v1.NodeList, error) {
	return s.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}
