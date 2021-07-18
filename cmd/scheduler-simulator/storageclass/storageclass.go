package storageclass

import (
	"context"
	"strings"

	"golang.org/x/xerrors"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages storageClasss.
type Service struct {
	client clientset.Interface
}

// NewStorageClassService initializes Service.
func NewStorageClassService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// Get returns the storageClass has given name.
// use simulatorID as namespace.
func (s *Service) Get(ctx context.Context, name string, simulatorID string) (*storagev1.StorageClass, error) {
	n, err := s.client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get storageClass: %w", err)
	}
	return n, nil
}

// List list all storageClass.
// use simulatorID as namespace.
func (s *Service) List(ctx context.Context, simulatorID string) (*storagev1.StorageClassList, error) {
	labelSelector := simulatorIDLabelSelector(simulatorID)
	pl, err := s.client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(labelSelector)})
	if err != nil {
		return nil, xerrors.Errorf("list storageClasss: %w", err)
	}
	return pl, nil
}

// Apply applies the storageClass.
// use simulatorID as namespace.
func (s *Service) Apply(ctx context.Context, simulatorID string, storageClass *v1.StorageClassApplyConfiguration) error {
	storageClass.WithKind("StorageClass")
	storageClass.WithAPIVersion("storage.k8s.io/v1")

	// add information for manage all user's node.
	addSimulatorIDLabel(storageClass, simulatorID)
	addNameSuffix(storageClass, simulatorID)

	_, err := s.client.StorageV1().StorageClasses().Apply(ctx, storageClass, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return xerrors.Errorf("apply storageClass: %w", err)
	}

	return nil
}

// Delete deletes the storageClass has given name.
// use simulatorID as namespace.
func (s *Service) Delete(ctx context.Context, name string, simulatorID string) error {
	err := s.client.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return xerrors.Errorf("delete storageClass: %w", err)
	}

	return nil
}

// addNameSuffix adds suffix to name
// Make storageClass names unique by using simulatorID as a suffix.
func addNameSuffix(sac *v1.StorageClassApplyConfiguration, suffix string) *v1.StorageClassApplyConfiguration {
	if sac == nil || sac.Name == nil {
		return sac
	}
	if strings.HasSuffix(*sac.Name, suffix) {
		return sac
	}

	// Add the suffix to the name only if the name don't have the suffix.
	newName := *sac.Name + "-" + suffix
	sac.Name = &newName
	return sac
}

const simulatorIDLabelKey = "simulatorID"

// addSimulatorIDLabel adds simulatorID to label.
func addSimulatorIDLabel(sac *v1.StorageClassApplyConfiguration, simulatorID string) *v1.StorageClassApplyConfiguration {
	if sac == nil {
		return sac
	}
	if sac.Labels == nil {
		sac.Labels = map[string]string{}
	}
	sac.Labels[simulatorIDLabelKey] = simulatorID
	return sac
}

func simulatorIDLabelSelector(simulatorID string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			simulatorIDLabelKey: simulatorID,
		},
	}
}
