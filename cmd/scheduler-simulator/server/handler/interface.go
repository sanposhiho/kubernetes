package handler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// NodeService represents service for manage Nodes.
type NodeService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.Node, error)
	List(ctx context.Context, simulatorID string) (*corev1.NodeList, error)
	Apply(ctx context.Context, simulatorID string, node *v1.NodeApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.Pod, error)
	List(ctx context.Context, simulatorID string) (*corev1.PodList, error)
	Apply(ctx context.Context, simulatorID string, pod *v1.PodApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// NamespaceService represents service for manage Namespaces.
type NamespaceService interface {
	Apply(ctx context.Context, namespace *v1.NamespaceApplyConfiguration) (*corev1.Namespace, error)
}
