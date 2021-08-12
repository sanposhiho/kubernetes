package di

import (
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/sanposhiho/k8s-scheduler-simulator/node"
	"github.com/sanposhiho/k8s-scheduler-simulator/persistentvolume"
	"github.com/sanposhiho/k8s-scheduler-simulator/persistentvolumeclaim"
	"github.com/sanposhiho/k8s-scheduler-simulator/pod"
	"github.com/sanposhiho/k8s-scheduler-simulator/scheduler"
	"github.com/sanposhiho/k8s-scheduler-simulator/storageclass"
)

// Container saves dependencies for handler.
type Container struct {
	nodeService         NodeService
	podService          PodService
	pvService           PersistentVolumeService
	pvcService          PersistentVolumeClaimService
	storageClassService StorageClassService
	schedulerService    SchedulerService
}

// NewDIContainer initializes Container.
func NewDIContainer(client clientset.Interface, restclientCfg *restclient.Config) *Container {
	c := &Container{}

	// initialize each service
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg)
	c.podService = pod.NewPodService(client)
	c.nodeService = node.NewNodeService(client, c.podService)

	return c
}

// NodeService returns handler.NodeService.
func (c *Container) NodeService() NodeService {
	return c.nodeService
}

// PodService returns handler.PodService.
func (c *Container) PodService() PodService {
	return c.podService
}

// StorageClassService returns handler.StorageClassService.
func (c *Container) StorageClassService() StorageClassService {
	return c.storageClassService
}

// PersistentVolumeService returns handler.PersistentVolumeService.
func (c *Container) PersistentVolumeService() PersistentVolumeService {
	return c.pvService
}

// PersistentVolumeClaimService returns handler.PersistentVolumeClaimService.
func (c *Container) PersistentVolumeClaimService() PersistentVolumeClaimService {
	return c.pvcService
}

func (c *Container) SchedulerService() SchedulerService {
	return c.schedulerService
}
