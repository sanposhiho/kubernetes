package persistentvolumeclaim

import (
	"context"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages persistentVolumeClaims.
type Service struct {
	client clientset.Interface
}

// NewPersistentVolumeClaimService initializes Service.
func NewPersistentVolumeClaimService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// Get returns the persistentVolumeClaims has given name.
// use simulatorID as namespace.
func (s *Service) Get(ctx context.Context, name string, simulatorID string) (*corev1.PersistentVolumeClaim, error) {
	n, err := s.client.CoreV1().PersistentVolumeClaims(simulatorID).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get persistentVolumeClaim: %w", err)
	}
	return n, nil
}

// List list all persistentVolumeClaims.
// use simulatorID as namespace.
func (s *Service) List(ctx context.Context, simulatorID string) (*corev1.PersistentVolumeClaimList, error) {
	pl, err := s.client.CoreV1().PersistentVolumeClaims(simulatorID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list persistentVolumeClaims: %w", err)
	}
	return pl, nil
}

// Apply applies the persistentVolumeClaims.
// use simulatorID as namespace.
func (s *Service) Apply(ctx context.Context, simulatorID string, persistentVolumeClaime *v1.PersistentVolumeClaimApplyConfiguration) error {
	persistentVolumeClaime.WithKind("PersistentVolumeClaim")
	persistentVolumeClaime.WithAPIVersion("v1")

	_, err := s.client.CoreV1().PersistentVolumeClaims(simulatorID).Apply(ctx, persistentVolumeClaime, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return xerrors.Errorf("apply persistentVolumeClaim: %w", err)
	}

	return nil
}

// Delete deletes the persistentVolumeClaims has given name.
// use simulatorID as namespace.
func (s *Service) Delete(ctx context.Context, name string, simulatorID string) error {
	err := s.client.CoreV1().PersistentVolumeClaims(simulatorID).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return xerrors.Errorf("delete persistentVolumeClaim: %w", err)
	}

	return nil
}
