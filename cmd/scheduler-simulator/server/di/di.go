package di

import (
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/kubernetes/cmd/scheduler-simulator/node"
	"k8s.io/kubernetes/cmd/scheduler-simulator/pod"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/handler"
)

type Container struct {
	nodeService handler.NodeService
	podService  handler.PodService
}

func NewDIContainer(client clientset.Interface, podInformer coreinformers.PodInformer) *Container {
	c := &Container{}

	// initialize each service
	c.nodeService = node.NewNodeService(client)
	c.podService = pod.NewPodService(client, podInformer)

	return c
}

func (c *Container) NodeService() handler.NodeService {
	return c.nodeService
}

func (c *Container) PodService() handler.PodService {
	return c.podService
}
