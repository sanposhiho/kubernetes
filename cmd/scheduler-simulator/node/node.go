package node

//go:generate mockgen -destination=./mock_$GOPACKAGE/$GOFILE --build_flags=--mod=mod . PodService
//go:generate mockgen -destination=./mock_clientset/clientset.go --build_flags=--mod=mod k8s.io/client-go/kubernetes Interface
// --build_flags is need for https://github.com/golang/mock#reflect-vendoring-error

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages node.
type Service struct {
	client     clientset.Interface
	podService PodService
}

// PodService represents service for manage Pods.
type PodService interface {
	List(ctx context.Context) (*corev1.PodList, error)
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
func (s *Service) Get(ctx context.Context, name string) (*corev1.Node, error) {
	n, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get nodes: %w", err)
	}
	return n, nil
}

// List lists all nodes.
func (s *Service) List(ctx context.Context) (*corev1.NodeList, error) {
	nl, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list nodes: %w", err)
	}
	return nl, nil
}

// Apply applies nodes.
func (s *Service) Apply(ctx context.Context, nac *v1.NodeApplyConfiguration) (*corev1.Node, error) {
	newNode, err := s.client.CoreV1().Nodes().Apply(ctx, nac, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply node: %w", err)
	}

	return newNode, nil
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
