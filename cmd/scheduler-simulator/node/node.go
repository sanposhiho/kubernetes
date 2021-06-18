package node

//go:generate mockgen -destination=./mock_$GOPACKAGE/$GOFILE --build_flags=--mod=mod . PodService
//go:generate mockgen -destination=./mock_clientset/clientset.go --build_flags=--mod=mod k8s.io/client-go/kubernetes Interface
// --build_flags is need for https://github.com/golang/mock#reflect-vendoring-error

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages node
type Service struct {
	client     clientset.Interface
	podService PodService
}

// PodService represents service for manage Pods.
type PodService interface {
	List(ctx context.Context) (*v1.PodList, error)
	Delete(ctx context.Context, name string) error
}

// NewNodeService initializes Service.
func NewNodeService(client clientset.Interface, ps PodService) *Service {
	return &Service{
		client:     client,
		podService: ps,
	}
}

// Get returns the node has given name.
func (s *Service) Get(ctx context.Context, name string) (*v1.Node, error) {
	n, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get nodes: %w", err)
	}
	return n, nil
}

// List lists all nodes.
func (s *Service) List(ctx context.Context) (*v1.NodeList, error) {
	nl, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list nodes: %w", err)
	}
	return nl, nil
}

// Create creates a node.
func (s *Service) Create(ctx context.Context) (*v1.Node, error) {
	// TODO: users specify specs.
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

	node, err := s.client.CoreV1().Nodes().Create(ctx, baseNode, metav1.CreateOptions{})
	if err != nil {
		return nil, xerrors.Errorf("create node: %w", err)
	}

	return node, nil
}

// Delete deletes the node has given name.
func (s *Service) Delete(ctx context.Context, name string) error {
	pl, err := s.podService.List(ctx)
	if err != nil {
		return xerrors.Errorf("list pods: %w", err)
	}

	// delete pods on node
	var eg errgroup.Group
	for i := range pl.Items {
		pod := pl.Items[i]
		if name != pod.Spec.NodeName {
			continue
		}
		eg.Go(func() error {
			if err := s.podService.Delete(ctx, pod.Name); err != nil {
				return xerrors.Errorf("delete pod: %w", err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("delete pods on node: %w", err)
	}

	// delete node
	if err := s.client.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return xerrors.Errorf("delete node: %w", err)
	}
	return nil
}
