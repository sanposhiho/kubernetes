package di

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	configv1 "k8s.io/client-go/applyconfigurations/core/v1"
	storageconfigv1 "k8s.io/client-go/applyconfigurations/storage/v1"

	"k8s.io/kubernetes/pkg/scheduler/apis/config"
)

// NamespaceService represents service for manage Namespaces.
type NamespaceService interface {
	Apply(ctx context.Context, namespace *configv1.NamespaceApplyConfiguration) (*corev1.Namespace, error)
}

// NodeService represents service for manage Nodes.
type NodeService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.Node, error)
	List(ctx context.Context, simulatorID string) (*corev1.NodeList, error)
	Apply(ctx context.Context, simulatorID string, node *configv1.NodeApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// PersistentVolumeService represents service for manage Pods.
type PersistentVolumeService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.PersistentVolume, error)
	List(ctx context.Context, simulatorID string) (*corev1.PersistentVolumeList, error)
	Apply(ctx context.Context, simulatorID string, pv *configv1.PersistentVolumeApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// PersistentVolumeClaimService represents service for manage Nodes.
type PersistentVolumeClaimService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.PersistentVolumeClaim, error)
	List(ctx context.Context, simulatorID string) (*corev1.PersistentVolumeClaimList, error)
	Apply(ctx context.Context, simulatorID string, pvc *configv1.PersistentVolumeClaimApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.Pod, error)
	List(ctx context.Context, simulatorID string) (*corev1.PodList, error)
	Apply(ctx context.Context, simulatorID string, pod *configv1.PodApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// SchedulerService represents service for manage scheduler.
type SchedulerService interface {
	GetSchedulerConfig() *config.KubeSchedulerConfiguration
	RestartScheduler(cfg *config.KubeSchedulerConfiguration) error
	StartScheduler(cfg *config.KubeSchedulerConfiguration) error
	ShutdownScheduler()
}

// StorageClassService represents service for manage Pods.
type StorageClassService interface {
	Get(ctx context.Context, name string, simulatorID string) (*storagev1.StorageClass, error)
	List(ctx context.Context, simulatorID string) (*storagev1.StorageClassList, error)
	Apply(ctx context.Context, simulatorID string, sc *storageconfigv1.StorageClassApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}
