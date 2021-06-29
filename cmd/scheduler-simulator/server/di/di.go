package di

import (
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/kubernetes/cmd/scheduler-simulator/namespace"
	"k8s.io/kubernetes/cmd/scheduler-simulator/node"
	"k8s.io/kubernetes/cmd/scheduler-simulator/pod"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/handler"
)

// Container saves dependencies for handler.
type Container struct {
	nodeService      handler.NodeService
	podService       handler.PodService
	namespaceService handler.NamespaceService
}

// NewDIContainer initializes Container.
func NewDIContainer(client clientset.Interface, podInformer coreinformers.PodInformer) *Container {
	c := &Container{}

	// initialize each service
	c.podService = pod.NewPodService(client, podInformer)
	c.nodeService = node.NewNodeService(client, c.podService)
	c.namespaceService = namespace.NewNamespaceService(client)

	return c
}

// NodeService returns handler.NodeService.
func (c *Container) NodeService() handler.NodeService {
	return c.nodeService
}

// PodService returns handler.PodService.
func (c *Container) PodService() handler.PodService {
	return c.podService
}

// NamespaceService returns handler.NamespaceService.
func (c *Container) NamespaceService() handler.NamespaceService {
	return c.namespaceService
}
