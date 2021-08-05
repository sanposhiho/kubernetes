package scheduler

import (
	"context"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/util"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/filter"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/score"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
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
		return xerrors.Errorf("start scheduler: %w", err)
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

	s.currentSchedulerCfg = cfg.DeepCopy()

	cfg, err := convertConfigurationForSimulator(cfg)
	if err != nil {
		cancel()
		return xerrors.Errorf("convert scheduler config to apply: %w", err)
	}

	// TODO: error handling after refactoring
	registry, _ := plugin.NewRegistry(informerFactory, clientSet)

	sched, err := scheduler.New(
		clientSet,
		informerFactory,
		profile.NewRecorderFactory(evtBroadcaster),
		ctx.Done(),
		scheduler.WithKubeConfig(restConfig),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithPercentageOfNodesToScore(cfg.PercentageOfNodesToScore),
		scheduler.WithPodMaxBackoffSeconds(cfg.PodMaxBackoffSeconds),
		scheduler.WithPodInitialBackoffSeconds(cfg.PodInitialBackoffSeconds),
		scheduler.WithExtenders(cfg.Extenders...),
		scheduler.WithParallelism(cfg.Parallelism),
		scheduler.WithFrameworkOutOfTreeRegistry(registry),
	)
	if err != nil {
		cancel()
		return xerrors.Errorf("create scheduler: %w", err)
	}

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	go sched.Run(ctx)

	s.shutdownfn = cancel

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

// convertConfigurationForSimulator convert KubeSchedulerConfiguration to apply scheduler on simulator
// (1) It excludes non-allowed changes. Now, we accept only the change of Profiles field.
// (2) It replaces filter/score default-plugins with plugin for simulator.
func convertConfigurationForSimulator(cfg *config.KubeSchedulerConfiguration) (*config.KubeSchedulerConfiguration, error) {
	newcfg := cfg.DeepCopy()
	if len(newcfg.Profiles) == 0 {
		newcfg.Profiles = []config.KubeSchedulerProfile{
			{
				SchedulerName: v1.DefaultSchedulerName,
				Plugins:       &config.Plugins{},
			},
		}
	}

	pc, err := plugin.NewPluginConfig()
	if err != nil {
		return nil, xerrors.Errorf("get plugin configs: %w", err)
	}

	for i := range newcfg.Profiles {
		if newcfg.Profiles[i].Plugins == nil {
			newcfg.Profiles[i].Plugins = &config.Plugins{}
		}

		// TODO: error handling after refactoring
		newcfg.Profiles[i].Plugins.Score.Enabled, _ = score.PluginsForSimulator(newcfg.Profiles[i].Plugins.Score.Disabled)
		newcfg.Profiles[i].Plugins.Score.Disabled, _ = score.DefaultScorePlugins()
		newcfg.Profiles[i].Plugins.Filter.Enabled, _ = filter.PluginsForSimulator(newcfg.Profiles[i].Plugins.Filter.Disabled)
		newcfg.Profiles[i].Plugins.Filter.Disabled, _ = filter.DefaultFilterPlugins()

		// TODO: support user custom plugin config
		newcfg.Profiles[i].PluginConfig = pc
	}

	defaultCfg, err := util.DefaultSchedulerConfig()
	if err != nil {
		return nil, xerrors.Errorf("get default scheduler config: %w", err)
	}
	// set default value to all field other than Profiles.
	defaultCfg.Profiles = newcfg.Profiles
	newcfg = defaultCfg

	return newcfg, nil
}
