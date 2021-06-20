package handler

import (
	"context"

	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	corev1 "k8s.io/api/core/v1"
)

// NodeService represents service for manage Nodes.
type NodeService interface {
	Get(ctx context.Context, name string) (*corev1.Node, error)
	List(ctx context.Context) (*corev1.NodeList, error)
	Apply(ctx context.Context, node *v1.NodeApplyConfiguration) (*corev1.Node, error)
	Delete(ctx context.Context, name string) error
}

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string) (*corev1.Pod, error)
	List(ctx context.Context) (*corev1.PodList, error)
	Apply(ctx context.Context, pod *v1.PodApplyConfiguration) (*corev1.Pod, error)
	Delete(ctx context.Context, name string) error
}
