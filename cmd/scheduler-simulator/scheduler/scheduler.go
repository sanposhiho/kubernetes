package scheduler

import (
	"context"

	"k8s.io/klog/v2"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	kubeschedulerscheme "k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/profile"
)

// Service manages scheduler.
type Service struct {
	// function to shutdown scheduler.
	shutdownfn func()

	clientset           clientset.Interface
	restclientCfg       *restclient.Config
	currentSchedulerCfg *config.KubeSchedulerConfiguration
}

// NewSchedulerService starts scheduler and return *Service.
func NewSchedulerService(client clientset.Interface, restclientCfg *restclient.Config) *Service {
	return &Service{clientset: client, restclientCfg: restclientCfg}
}

func (s *Service) RestartScheduler(cfg *config.KubeSchedulerConfiguration) error {
	s.ShutdownScheduler()

	if err := s.StartScheduler(cfg); err != nil {
		return xerrors.Errorf("start scheduler: %w")
	}
	return nil
}

// StartScheduler starts scheduler.
func (s *Service) StartScheduler(cfg *config.KubeSchedulerConfiguration) error {
	clientSet := s.clientset
	restConfig := s.restclientCfg
	ctx, cancel := context.WithCancel(context.Background())

	informerFactory := scheduler.NewInformerFactory(clientSet, 0)
	evtBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{
		Interface: clientSet.EventsV1(),
	})

	evtBroadcaster.StartRecordingToSink(ctx.Done())

	sched, err := scheduler.New(
		clientSet,
		informerFactory,
		profile.NewRecorderFactory(evtBroadcaster),
		ctx.Done(),
		scheduler.WithKubeConfig(restConfig),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithAlgorithmSource(cfg.AlgorithmSource),
		scheduler.WithPercentageOfNodesToScore(cfg.PercentageOfNodesToScore),
		scheduler.WithPodMaxBackoffSeconds(cfg.PodMaxBackoffSeconds),
		scheduler.WithPodInitialBackoffSeconds(cfg.PodInitialBackoffSeconds),
		scheduler.WithExtenders(cfg.Extenders...),
		scheduler.WithParallelism(cfg.Parallelism),
		scheduler.WithFrameworkOutOfTreeRegistry(plugin.NewRegistry(informerFactory, clientSet)),
	)
	if err != nil {
		cancel()
		return xerrors.Errorf("create scheduler: %w", err)
	}

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	go sched.Run(ctx)

	s.shutdownfn = cancel
	s.currentSchedulerCfg = cfg

	return nil
}

func (s *Service) ShutdownScheduler() {
	if s.shutdownfn != nil {
		klog.Info("shutdown scheduler...")
		s.shutdownfn()
	}
}

func (s *Service) GetSchedulerConfig() *config.KubeSchedulerConfiguration {
	return s.currentSchedulerCfg
}

// DefaultConfig creates KubeSchedulerConfiguration default configuration.
func DefaultConfig() (*config.KubeSchedulerConfiguration, error) {
	gvk := v1beta1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration")
	cfg := config.KubeSchedulerConfiguration{}
	_, _, err := kubeschedulerscheme.Codecs.UniversalDecoder().Decode(nil, &gvk, &cfg)
	if err != nil {
		return nil, xerrors.Errorf("decode config: %w", err)
	}

	plugins := plugin.NewPlugins()

	pc, err := plugin.NewPluginConfig()
	if err != nil {
		return nil, xerrors.Errorf("get plugin configs: %w", err)
	}

	cfg.Profiles = []config.KubeSchedulerProfile{
		{
			SchedulerName: v1.DefaultSchedulerName,
			Plugins:       plugins,
			PluginConfig:  pc,
		},
	}
	return &cfg, nil

}

func SwitchPluginsForSimulator(cfg *config.KubeSchedulerConfiguration) *config.KubeSchedulerConfiguration {

}

// OnlyProfile exclude all but Profiles fields.
func OnlyProfile(cfg *config.KubeSchedulerConfiguration) *config.KubeSchedulerConfiguration {
	return &config.KubeSchedulerConfiguration{Profiles: cfg.Profiles}
}
