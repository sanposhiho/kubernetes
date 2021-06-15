package handler

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type NodeService interface {
	Create(ctx context.Context) (*v1.Node, error)
	List(ctx context.Context) (*v1.NodeList, error)
}

type PodService interface {
	Create(ctx context.Context) (*v1.Pod, error)
	List(ctx context.Context) (*v1.PodList, error)
}
