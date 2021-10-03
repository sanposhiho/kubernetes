package mini_kube_scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/docker/go-metrics"
	restclient "k8s.io/client-go/rest"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"

	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/dynamic/dynamicinformer"
	frameworkplugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	internalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"golang.org/x/xerrors"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/profile"
)

// Option configures a Scheduler
type Option func(*schedulerOptions)

// WithComponentConfigVersion sets the component config version to the
// KubeSchedulerConfiguration version used. The string should be the full
// scheme group/version of the external type we converted from (for example
// "kubescheduler.config.k8s.io/v1beta2")
func WithComponentConfigVersion(apiVersion string) Option {
	return func(o *schedulerOptions) {
		o.componentConfigVersion = apiVersion
	}
}

// WithKubeConfig sets the kube config for Scheduler.
func WithKubeConfig(cfg *restclient.Config) Option {
	return func(o *schedulerOptions) {
		o.kubeConfig = cfg
	}
}

// WithProfiles sets profiles for Scheduler. By default, there is one profile
// with the name "default-scheduler".
func WithProfiles(p ...schedulerapi.KubeSchedulerProfile) Option {
	return func(o *schedulerOptions) {
		o.profiles = p
		o.applyDefaultProfile = false
	}
}

// WithParallelism sets the parallelism for all scheduler algorithms. Default is 16.
func WithParallelism(threads int32) Option {
	return func(o *schedulerOptions) {
		o.parallelism = threads
	}
}

// WithLegacyPolicySource sets legacy policy config file source.
func WithLegacyPolicySource(source *schedulerapi.SchedulerPolicySource) Option {
	return func(o *schedulerOptions) {
		o.legacyPolicySource = source
	}
}

// WithPercentageOfNodesToScore sets percentageOfNodesToScore for Scheduler, the default value is 50
func WithPercentageOfNodesToScore(percentageOfNodesToScore int32) Option {
	return func(o *schedulerOptions) {
		o.percentageOfNodesToScore = percentageOfNodesToScore
	}
}

// WithFrameworkOutOfTreeRegistry sets the registry for out-of-tree plugins. Those plugins
// will be appended to the default registry.
func WithFrameworkOutOfTreeRegistry(registry frameworkruntime.Registry) Option {
	return func(o *schedulerOptions) {
		o.frameworkOutOfTreeRegistry = registry
	}
}

// WithPodInitialBackoffSeconds sets podInitialBackoffSeconds for Scheduler, the default value is 1
func WithPodInitialBackoffSeconds(podInitialBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.podInitialBackoffSeconds = podInitialBackoffSeconds
	}
}

// WithPodMaxBackoffSeconds sets podMaxBackoffSeconds for Scheduler, the default value is 10
func WithPodMaxBackoffSeconds(podMaxBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.podMaxBackoffSeconds = podMaxBackoffSeconds
	}
}

// WithExtenders sets extenders for the Scheduler
func WithExtenders(e ...schedulerapi.Extender) Option {
	return func(o *schedulerOptions) {
		o.extenders = e
	}
}

// FrameworkCapturer is used for registering a notify function in building framework.
type FrameworkCapturer func(schedulerapi.KubeSchedulerProfile)

// WithBuildFrameworkCapturer sets a notify function for getting buildFramework details.
func WithBuildFrameworkCapturer(fc FrameworkCapturer) Option {
	return func(o *schedulerOptions) {
		o.frameworkCapturer = fc
	}
}

// New returns a Scheduler
func New(client clientset.Interface,
	informerFactory informers.SharedInformerFactory,
	recorderFactory profile.RecorderFactory,
	stopCh <-chan struct{},
	opts ...Option) (*Scheduler, error) {

	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.applyDefaultProfile {
		var versionedCfg v1beta2.KubeSchedulerConfiguration
		scheme.Scheme.Default(&versionedCfg)
		cfg := config.KubeSchedulerConfiguration{}
		if err := scheme.Scheme.Convert(&versionedCfg, &cfg, nil); err != nil {
			return nil, err
		}
		options.profiles = cfg.Profiles
	}
	schedulerCache := internalcache.New(30*time.Second, stopEverything)

	registry := frameworkplugins.NewInTreeRegistry()
	if err := registry.Merge(options.frameworkOutOfTreeRegistry); err != nil {
		return nil, err
	}

	snapshot := internalcache.NewEmptySnapshot()
	clusterEventMap := make(map[framework.ClusterEvent]sets.String)

	configurator := &Configurator{
		componentConfigVersion:   options.componentConfigVersion,
		client:                   client,
		kubeConfig:               options.kubeConfig,
		recorderFactory:          recorderFactory,
		informerFactory:          informerFactory,
		schedulerCache:           schedulerCache,
		StopEverything:           stopEverything,
		percentageOfNodesToScore: options.percentageOfNodesToScore,
		podInitialBackoffSeconds: options.podInitialBackoffSeconds,
		podMaxBackoffSeconds:     options.podMaxBackoffSeconds,
		profiles:                 append([]schedulerapi.KubeSchedulerProfile(nil), options.profiles...),
		registry:                 registry,
		nodeInfoSnapshot:         snapshot,
		extenders:                options.extenders,
		frameworkCapturer:        options.frameworkCapturer,
		parallellism:             options.parallelism,
		clusterEventMap:          clusterEventMap,
	}

	metrics.Register()

	var sched *Scheduler
	// Create the config from component config
	sc, err := configurator.create()
	if err != nil {
		return nil, fmt.Errorf("couldn't create scheduler: %v", err)
	}
	sched = sc

	// Additional tweaks to the config produced by the configurator.
	sched.StopEverything = stopEverything
	sched.client = client

	// Build dynamic client and dynamic informer factory
	var dynInformerFactory dynamicinformer.DynamicSharedInformerFactory
	// options.kubeConfig can be nil in tests.
	if options.kubeConfig != nil {
		dynClient := dynamic.NewForConfigOrDie(options.kubeConfig)
		dynInformerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynClient, 0, v1.NamespaceAll, nil)
	}

	addAllEventHandlers(sched, informerFactory, dynInformerFactory, unionedGVKs(clusterEventMap))
	return sched, nil
}

func (s *Scheduler) Run(ctx context.Context) {
	s.SchedulingQueue.Run()
	wait.UntilWithContext(ctx, s.scheduleOne, 0)
	s.SchedulingQueue.Close()
}

func (s *Scheduler) scheduleOne(ctx context.Context) {
	podInfo, err := s.NextPod()
	if err != nil {
		klog.Errorf("failed to get next pod: %w", err)
	}

	state := framework.NewCycleState()
	result, err := s.Schedule(ctx, state, podInfo.Pod)
	if err != nil {
		//TODO
		fmt.Println("FAIL!")
		fmt.Println(err)
		return
	}
	fmt.Println("SUCCESS!")
	fmt.Println(result)
	return
}

func (s *Scheduler) NextPod() (*framework.QueuedPodInfo, error) {
	podInfo, err := s.SchedulingQueue.Pop()
	if err != nil {
		return nil, xerrors.Errorf("pop pod info from queue: %w", err)
	}
	return podInfo, nil
}

func (s *Scheduler) Schedule(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (scheduleResult *ScheduleResult, err error) {
	// pre filter
	status := s.RunPreFilterPlugins(ctx, state, pod)
	if !status.IsSuccess() {
		return nil, status.AsError()
	}

	// get nodes
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	// filter
	fasibleNodes, status := s.RunFilterPlugins(ctx, state, pod, nodes.Items)
	if !status.IsSuccess() {
		return nil, status.AsError()
	}

	// pre score
	status = s.RunPreScorePlugins(ctx, state, pod, fasibleNodes)
	if !status.IsSuccess() {
		return nil, status.AsError()
	}

	// score
	scoreList, status := s.RunScorePlugins(ctx, state, pod, fasibleNodes)
	if !status.IsSuccess() {
		return nil, status.AsError()
	}

	host, err := selectHost(scoreList)
	if err != nil {
		return nil, fmt.Errorf("select host: %w", err)
	}

	return &ScheduleResult{SuggestedHost: host}, nil
}

func (s *Scheduler) createPluginToNodeScores(nodes []*v1.Node) framework.PluginToNodeScores {
	pluginToNodeScores := make(framework.PluginToNodeScores, len(s.scorePlugins))
	for _, pl := range s.scorePlugins {
		pluginToNodeScores[pl.Name()] = make(framework.NodeScoreList, len(nodes))
	}

	return pluginToNodeScores
}

func (s *Scheduler) RunPreFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	for _, pl := range s.preFilterPlugins {
		status := pl.PreFilter(ctx, state, pod)
		if !status.IsSuccess() {
			status.SetFailedPlugin(pl.Name())
			if status.IsUnschedulable() {
				return status
			}

			return framework.AsStatus(fmt.Errorf("running PreFilter plugin %q: %w", pl.Name(), status.AsError())).WithFailedPlugin(pl.Name())
		}
	}

	return nil
}

func (s *Scheduler) RunFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []v1.Node) ([]*v1.Node, *framework.Status) {
	feasibleNodes := make([]*v1.Node, len(nodes))

	// TODO: consider about nominated pod
	statuses := make(framework.PluginToStatus)
	for _, n := range nodes {
		nodeInfo := framework.NewNodeInfo()
		nodeInfo.SetNode(&n)
		for _, pl := range s.filterPlugins {
			status := pl.Filter(ctx, state, pod, nodeInfo)
			if !status.IsSuccess() {
				status.SetFailedPlugin(pl.Name())
				statuses[pl.Name()] = status
				return nil, statuses.Merge()
			}
			feasibleNodes = append(feasibleNodes, nodeInfo.Node())
		}
	}

	return feasibleNodes, statuses.Merge()
}

func (s *Scheduler) RunPreScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	for _, pl := range s.preScorePlugins {
		status := pl.PreScore(ctx, state, pod, nodes)
		if !status.IsSuccess() {
			return status
		}
	}

	return nil
}

func (s *Scheduler) RunScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) (framework.NodeScoreList, *framework.Status) {
	scoresMap := s.createPluginToNodeScores(nodes)

	// TODO: consider about nominated pod
	for index, n := range nodes {
		for _, pl := range s.scorePlugins {
			score, status := pl.Score(ctx, state, pod, n.Name)
			if !status.IsSuccess() {
				return nil, status
			}
			scoresMap[pl.Name()][index] = framework.NodeScore{
				Name:  n.Name,
				Score: score,
			}

			if pl.ScoreExtensions() != nil {
				status := pl.ScoreExtensions().NormalizeScore(ctx, state, pod, scoresMap[pl.Name()])
				if !status.IsSuccess() {
					return nil, status
				}
			}
		}
	}

	result := make(framework.NodeScoreList, 0, len(nodes))

	for i := range nodes {
		result = append(result, framework.NodeScore{Name: nodes[i].Name, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	return result, nil
}

func selectHost(nodeScoreList framework.NodeScoreList) (string, error) {
	if len(nodeScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.Name
			}
		}
	}
	return selected, nil
}
