package mini_kube_scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/scheduler"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	internalqueue "k8s.io/kubernetes/pkg/scheduler/internal/queue"
)

// ScheduleResult represents the result of one pod scheduled. It will contain
// the final selected Node, along with the selected intermediate information.
type ScheduleResult struct {
	// Name of the scheduler suggest host
	SuggestedHost string
	// Number of nodes scheduler evaluated on one pod scheduled
	EvaluatedNodes int
	// Number of feasible nodes on one pod scheduled
	FeasibleNodes int
}

// Scheduler is the same as k8s.io/kubernetes/pkg/scheduler.Scheduler.
type Scheduler struct {
	// It is expected that changes made via SchedulerCache will be observed
	// by NodeLister and Algorithm.
	SchedulerCache Cache

	Algorithm ScheduleAlgorithm

	Extenders []framework.Extender

	// Error is called if there is an error. It is passed the pod in
	// question, and the error
	Error func(*framework.QueuedPodInfo, error)

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	// SchedulingQueue holds pods to be scheduled
	SchedulingQueue internalqueue.SchedulingQueue

	client clientset.Interface

	preFilterPlugins []framework.PreFilterPlugin
	filterPlugins    []framework.FilterPlugin
	scorePlugins     []framework.ScorePlugin
	preScorePlugins  []framework.PreScorePlugin

	// plugin name is key
	scorePluginWeight map[string]int
}

// schedulerOptions is the same as k8s.io/kubernetes/pkg/scheduler.schedulerOptions.
type schedulerOptions struct {
	componentConfigVersion   string
	kubeConfig               *restclient.Config
	legacyPolicySource       *schedulerapi.SchedulerPolicySource
	percentageOfNodesToScore int32
	podInitialBackoffSeconds int64
	podMaxBackoffSeconds     int64
	// Contains out-of-tree plugins to be merged with the in-tree registry.
	frameworkOutOfTreeRegistry frameworkruntime.Registry
	profiles                   []schedulerapi.KubeSchedulerProfile
	extenders                  []schedulerapi.Extender
	frameworkCapturer          scheduler.FrameworkCapturer
	parallelism                int32
	applyDefaultProfile        bool
}

// defaultSchedulerOptions is the same as k8s.io/kubernetes/pkg/scheduler.defaultSchedulerOptions.
var defaultSchedulerOptions = schedulerOptions{
	//percentageOfNodesToScore: schedulerapi.DefaultPercentageOfNodesToScore,
	//podInitialBackoffSeconds: int64((1 * time.Second).Seconds()),
	//podMaxBackoffSeconds:     int64((10 * time.Second).Seconds()),
	//parallelism:              int32(16),
}

type Cache interface {
	// NodeCount returns the number of nodes in the cache.
	// DO NOT use outside of tests.
	NodeCount() int

	// PodCount returns the number of pods in the cache (including those from deleted nodes).
	// DO NOT use outside of tests.
	PodCount() (int, error)

	// AssumePod assumes a pod scheduled and aggregates the pod's information into its node.
	// The implementation also decides the policy to expire pod before being confirmed (receiving Add event).
	// After expiration, its information would be subtracted.
	AssumePod(pod *v1.Pod) error

	// FinishBinding signals that cache for assumed pod can be expired
	FinishBinding(pod *v1.Pod) error

	// ForgetPod removes an assumed pod from cache.
	ForgetPod(pod *v1.Pod) error

	// AddPod either confirms a pod if it's assumed, or adds it back if it's expired.
	// If added back, the pod's information would be added again.
	AddPod(pod *v1.Pod) error

	// UpdatePod removes oldPod's information and adds newPod's information.
	UpdatePod(oldPod, newPod *v1.Pod) error

	// RemovePod removes a pod. The pod's information would be subtracted from assigned node.
	RemovePod(pod *v1.Pod) error

	// GetPod returns the pod from the cache with the same namespace and the
	// same name of the specified pod.
	GetPod(pod *v1.Pod) (*v1.Pod, error)

	// IsAssumedPod returns true if the pod is assumed and not expired.
	IsAssumedPod(pod *v1.Pod) (bool, error)

	// AddNode adds overall information about node.
	// It returns a clone of added NodeInfo object.
	AddNode(node *v1.Node) *framework.NodeInfo

	// UpdateNode updates overall information about node.
	// It returns a clone of updated NodeInfo object.
	UpdateNode(oldNode, newNode *v1.Node) *framework.NodeInfo

	// RemoveNode removes overall information about node.
	RemoveNode(node *v1.Node) error

	// UpdateSnapshot updates the passed infoSnapshot to the current contents of Cache.
	// The node info contains aggregated information of pods scheduled (including assumed to be)
	// on this node.
	// The snapshot only includes Nodes that are not deleted at the time this function is called.
	// nodeinfo.Node() is guaranteed to be not nil for all the nodes in the snapshot.
	UpdateSnapshot(nodeSnapshot *Snapshot) error

	// Dump produces a dump of the current cache.
	Dump() *Dump
}

// Dump is a dump of the cache state.
type Dump struct {
	AssumedPods sets.String
	Nodes       map[string]*framework.NodeInfo
}
