package di

import (
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	node2 "k8s.io/kubernetes/cmd/scheduler-simulator/service/node"
	pod2 "k8s.io/kubernetes/cmd/scheduler-simulator/service/pod"

	"k8s.io/kubernetes/cmd/scheduler-simulator/server/handler"
)

type Container struct {
	nodeService handler.NodeService
	podService  handler.PodService
}

func NewDIContainer(client clientset.Interface, podInformer coreinformers.PodInformer) *Container {
	c := &Container{}

	// initialize each service
	c.podService = pod2.NewPodService(client, podInformer)
	c.nodeService = node2.NewNodeService(client, c.podService)

	return c
}

func (c *Container) NodeService() handler.NodeService {
	return c.nodeService
}

func (c *Container) PodService() handler.PodService {
	return c.podService
}
