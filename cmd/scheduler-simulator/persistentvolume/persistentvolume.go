package persistentvolume

import (
	"context"
	"strings"

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
	labelSelector := simulatorIDLabelSelector(simulatorID)
	pl, err := s.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(labelSelector)})
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

	// add information for manage all user's node.
	addSimulatorIDLabel(persistentVolume, simulatorID)
	addNameSuffix(persistentVolume, simulatorID)

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

// addNameSuffix adds suffix to name
// Make persistentVolume names unique by using simulatorID as a suffix.
func addNameSuffix(pvac *v1.PersistentVolumeApplyConfiguration, suffix string) *v1.PersistentVolumeApplyConfiguration {
	if pvac == nil || pvac.Name == nil {
		return pvac
	}
	if strings.HasSuffix(*pvac.Name, suffix) {
		return pvac
	}

	// Add the suffix to the name only if the name don't have the suffix.
	newName := *pvac.Name + "-" + suffix
	pvac.Name = &newName
	return pvac
}

const simulatorIDLabelKey = "simulatorID"

// addSimulatorIDLabel adds simulatorID to label.
func addSimulatorIDLabel(pvac *v1.PersistentVolumeApplyConfiguration, simulatorID string) *v1.PersistentVolumeApplyConfiguration {
	if pvac == nil {
		return pvac
	}
	if pvac.Labels == nil {
		pvac.Labels = map[string]string{}
	}
	pvac.Labels[simulatorIDLabelKey] = simulatorID
	return pvac
}

func simulatorIDLabelSelector(simulatorID string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			simulatorIDLabelKey: simulatorID,
		},
	}
}
