package di

import (
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/kubernetes/cmd/scheduler-simulator/node"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/handler"
)

type Container struct {
	nodeService handler.NodeService
}

func NewDIContainer(client clientset.Interface) *Container {
	c := &Container{}

	// initialize each service
	c.nodeService = node.NewNodeService(client)

	return c
}

func (c *Container) NodeService() handler.NodeService {
	return c.nodeService
}
