package di

import (
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/scheduler-simulator/etcd"
	"k8s.io/kubernetes/cmd/scheduler-simulator/pod"
	"k8s.io/kubernetes/cmd/scheduler-simulator/schedulerconfig"

	"k8s.io/kubernetes/cmd/scheduler-simulator/namespace"
	"k8s.io/kubernetes/cmd/scheduler-simulator/node"
	"k8s.io/kubernetes/cmd/scheduler-simulator/persistentvolume"
	"k8s.io/kubernetes/cmd/scheduler-simulator/persistentvolumeclaim"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/handler"
	"k8s.io/kubernetes/cmd/scheduler-simulator/storageclass"
)

// Container saves dependencies for handler.
type Container struct {
	nodeService            handler.NodeService
	podService             handler.PodService
	pvService              handler.PersistentVolumeService
	pvcService             handler.PersistentVolumeClaimService
	storageClassService    handler.StorageClassService
	namespaceService       handler.NamespaceService
	schedulerconfigService handler.SchedulerConfigService
}

// NewDIContainer initializes Container.
func NewDIContainer(client clientset.Interface, etcdclient *etcd.Client) *Container {
	c := &Container{}

	// initialize each service
	c.namespaceService = namespace.NewNamespaceService(client)
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerconfigService = schedulerconfig.NewSchedulerConfigService(etcdclient)
	c.podService = pod.NewPodService(client, c.schedulerconfigService)
	c.nodeService = node.NewNodeService(client, c.podService)

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

// StorageClassService returns handler.StorageClassService.
func (c *Container) StorageClassService() handler.StorageClassService {
	return c.storageClassService
}

// PersistentVolumeService returns handler.PersistentVolumeService.
func (c *Container) PersistentVolumeService() handler.PersistentVolumeService {
	return c.pvService
}

// PersistentVolumeClaimService returns handler.PersistentVolumeClaimService.
func (c *Container) PersistentVolumeClaimService() handler.PersistentVolumeClaimService {
	return c.pvcService
}

func (c *Container) SchedulerConfigService() handler.SchedulerConfigService {
	return c.schedulerconfigService
}
