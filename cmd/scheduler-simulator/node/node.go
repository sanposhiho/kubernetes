package node

//go:generate mockgen -destination=./mock_$GOPACKAGE/$GOFILE --build_flags=--mod=mod . PodService
//go:generate mockgen -destination=./mock_clientset/clientset.go --build_flags=--mod=mod k8s.io/client-go/kubernetes Interface
// --build_flags is need for https://github.com/golang/mock#reflect-vendoring-error

import (
	"context"
	"strings"

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
	List(ctx context.Context, simulatorID string) (*corev1.PodList, error)
	Delete(ctx context.Context, name string, simulatorID string) error
}

// NewNodeService initializes Service.
func NewNodeService(client clientset.Interface, ps PodService) *Service {
	return &Service{
		client:     client,
		podService: ps,
	}
}

// Get returns the node has given name.
func (s *Service) Get(ctx context.Context, name string, simulatorID string) (*corev1.Node, error) {
	n, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get nodes: %w", err)
	}

	return n, nil
}

// List lists all nodes.
func (s *Service) List(ctx context.Context, simulatorID string) (*corev1.NodeList, error) {
	labelSelector := simulatorIDLabelSelector(simulatorID)
	nl, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(labelSelector)})
	if err != nil {
		return nil, xerrors.Errorf("list nodes: %w", err)
	}

	return nl, nil
}

// Apply a unique node by using the simulator ID.
func (s *Service) Apply(ctx context.Context, simulatorID string, nac *v1.NodeApplyConfiguration) error {
	nac.WithAPIVersion("v1")
	nac.WithKind("Node")

	// add information for manage all user's node.
	addSimulatorIDLabel(nac, simulatorID)
	addNameSuffix(nac, simulatorID)

	_, err := s.client.CoreV1().Nodes().Apply(ctx, nac, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return xerrors.Errorf("apply node: %w", err)
	}

	return nil
}

// Delete deletes the node has given name.
func (s *Service) Delete(ctx context.Context, name string, simulatorID string) error {
	pl, err := s.podService.List(ctx, simulatorID)
	if err != nil {
		return xerrors.Errorf("list pods: %w", err)
	}

	// delete pods on node
	for i := range pl.Items {
		pod := pl.Items[i]
		if name != pod.Spec.NodeName {
			continue
		}

		if err := s.podService.Delete(ctx, pod.Name, simulatorID); err != nil {
			return xerrors.Errorf("delete pod: %w", err)
		}
	}

	// delete node
	if err := s.client.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return xerrors.Errorf("delete node: %w", err)
	}
	return nil
}

// addNameSuffix adds suffix to nac.Name
// Make Node names unique by using simulatorID as a suffix.
func addNameSuffix(nac *v1.NodeApplyConfiguration, suffix string) {
	if nac == nil || nac.Name == nil {
		return
	}
	if strings.HasSuffix(*nac.Name, suffix) {
		return
	}

	// Add the suffix to the name only if the name don't have the suffix.
	newName := *nac.Name + "-" + suffix
	nac.Name = &newName
}

const SimulatorIDLabelKey = "simulatorID"

// addSimulatorIDLabel adds simulatorID to label.
func addSimulatorIDLabel(nac *v1.NodeApplyConfiguration, simulatorID string) {
	if nac == nil {
		return
	}
	if nac.Labels == nil {
		nac.Labels = map[string]string{}
	}
	nac.Labels[SimulatorIDLabelKey] = simulatorID
}

func simulatorIDLabelSelector(simulatorID string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			SimulatorIDLabelKey: simulatorID,
		},
	}
}
