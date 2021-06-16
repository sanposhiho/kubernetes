package handler

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type NodeService interface {
	Get(ctx context.Context, name string) (*v1.Node, error)
	List(ctx context.Context) (*v1.NodeList, error)
	Create(ctx context.Context) (*v1.Node, error)
	Delete(ctx context.Context, name string) error
}

type PodService interface {
	Get(ctx context.Context, name string) (*v1.Pod, error)
	List(ctx context.Context) (*v1.PodList, error)
	Create(ctx context.Context) (*v1.Pod, error)
	Delete(ctx context.Context, name string) error
}
