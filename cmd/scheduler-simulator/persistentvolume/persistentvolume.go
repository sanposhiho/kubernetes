package persistentvolume

import (
	"context"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages persistentVolumes.
type Service struct {
	client clientset.Interface
}

// NewPersistentVolumeService initializes Service.
func NewPersistentVolumeService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// Get returns the persistentVolume has given name.
// use simulatorID as namespace.
func (s *Service) Get(ctx context.Context, name string, simulatorID string) (*corev1.PersistentVolume, error) {
	n, err := s.client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get persistentVolume: %w", err)
	}
	return n, nil
}

// List list all persistentVolumes.
// use simulatorID as namespace.
func (s *Service) List(ctx context.Context, simulatorID string) (*corev1.PersistentVolumeList, error) {
	pl, err := s.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list persistentVolumes: %w", err)
	}
	return pl, nil
}

// Apply applies the persistentVolume.
// use simulatorID as namespace.
func (s *Service) Apply(ctx context.Context, simulatorID string, persistentVolume *v1.PersistentVolumeApplyConfiguration) error {
	persistentVolume.WithKind("PersistentVolume")
	persistentVolume.WithAPIVersion("v1")

	_, err := s.client.CoreV1().PersistentVolumes().Apply(ctx, persistentVolume, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return xerrors.Errorf("apply persistentVolume: %w", err)
	}

	return nil
}

// Delete deletes the persistentVolume has given name.
// use simulatorID as namespace.
func (s *Service) Delete(ctx context.Context, name string, simulatorID string) error {
	err := s.client.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return xerrors.Errorf("delete persistentVolume: %w", err)
	}

	return nil
}
