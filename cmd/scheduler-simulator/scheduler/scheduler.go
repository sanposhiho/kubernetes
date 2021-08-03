package scheduler

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/filter"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/score"
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

	if err := SchedulerConfigurationForSimulator(cfg); err != nil {
		cancel()
		return xerrors.Errorf("convert scheduler config to apply: %w", err)
	}

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

	return nil
}

// DefaultSchedulerConfig creates KubeSchedulerConfiguration default configuration.
func DefaultSchedulerConfig() (*config.KubeSchedulerConfiguration, error) {
	gvk := v1beta1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration")
	cfg := config.KubeSchedulerConfiguration{}
	_, _, err := kubeschedulerscheme.Codecs.UniversalDecoder().Decode(nil, &gvk, &cfg)
	if err != nil {
		return nil, xerrors.Errorf("decode config: %w", err)
	}

	return &cfg, nil
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

// SchedulerConfigurationForSimulator convert KubeSchedulerConfiguration to apply scheduler on simulator
// (1) It excludes non-allowed changes. Now, we accept only the change of Profiles field.
// (2) It replaces filter/score default-plugins with plugin for simulator.
func SchedulerConfigurationForSimulator(cfg *config.KubeSchedulerConfiguration) error {
	if len(cfg.Profiles) == 0 {
		cfg.Profiles = []config.KubeSchedulerProfile{
			{
				SchedulerName: v1.DefaultSchedulerName,
				Plugins:       &config.Plugins{},
			},
		}
	}

	pc, err := plugin.NewPluginConfig()
	if err != nil {
		return xerrors.Errorf("get plugin configs: %w", err)
	}

	for i := range cfg.Profiles {
		if cfg.Profiles[i].Plugins == nil {
			cfg.Profiles[i].Plugins = &config.Plugins{}
		}

		cfg.Profiles[i].Plugins.Score.Enabled = score.ScorePlugins(cfg.Profiles[i].Plugins.Score.Disabled)
		cfg.Profiles[i].Plugins.Score.Disabled = score.DefaultScorePlugins()
		cfg.Profiles[i].Plugins.Filter.Enabled = filter.FilterPlugins(cfg.Profiles[i].Plugins.Filter.Disabled)
		cfg.Profiles[i].Plugins.Filter.Disabled = filter.DefaultFilterPlugins()

		// TODO: support user custom plugin config
		cfg.Profiles[i].PluginConfig = pc
	}

	defaultCfg, err := DefaultSchedulerConfig()
	if err != nil {
		return xerrors.Errorf("get default scheduler config: %w", err)
	}
	// set default value to all field other than Profiles.
	defaultCfg.Profiles = cfg.Profiles
	cfg = defaultCfg

	return nil
}
