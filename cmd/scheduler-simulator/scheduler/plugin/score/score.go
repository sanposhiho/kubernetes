package score

import (
	"context"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/util"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func NewRegistryForScoreRecord(s *schedulingresultstore.Store) (map[string]schedulerRuntime.PluginFactory, error) {
	ret := map[string]schedulerRuntime.PluginFactory{}
	rs := plugins.NewInTreeRegistry()

	defaultpls, err := DefaultScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	for _, pl := range defaultpls {
		pl := pl
		r := rs[pl.Name]
		factory := func(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
			p, err := r(configuration, f)
			if err != nil {
				return nil, err
			}
			typed, ok := p.(framework.ScorePlugin)
			if !ok {
				return nil, xerrors.New("not score plugin")
			}
			return NewScoreRecordPlugin(s, typed, pl.Weight), nil
		}
		ret[pluginName(pl.Name)] = factory
	}

	return ret, nil
}

func DefaultScorePlugins() ([]config.Plugin, error) {
	defaultConfig, err := util.DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return nil, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Score.Enabled, nil
}

// PluginsForSimulator create scorePluginForSimulator for simulator.
func PluginsForSimulator(disabledPlugins []config.Plugin) ([]config.Plugin, error) {
	// true means the plugin is disabled
	disabledMap := map[string]bool{}
	for _, p := range disabledPlugins {
		disabledMap[p.Name] = true
	}

	defaultpls, err := DefaultScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	var ret []config.Plugin
	for _, dp := range defaultpls {
		if !disabledMap[dp.Name] {
			ret = append(ret, config.Plugin{Name: pluginName(dp.Name), Weight: dp.Weight})
		}
	}
	return ret, nil
}

func PluginConfigs() ([]config.PluginConfig, error) {
	defaultpls, err := DefaultScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	defaultcfg, err := util.DefaultSchedulerConfig()
	if err != nil || len(defaultcfg.Profiles) != 1 {
		return nil, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	pluginConfig := make(map[string]runtime.Object, len(defaultcfg.Profiles[0].PluginConfig))
	for i := range defaultcfg.Profiles[0].PluginConfig {
		name := defaultcfg.Profiles[0].PluginConfig[i].Name
		pluginConfig[name] = defaultcfg.Profiles[0].PluginConfig[i].Args
	}

	ret := make([]config.PluginConfig, len(defaultpls))
	for i, p := range defaultpls {
		name := p.Name

		ret[i] = config.PluginConfig{
			Name: pluginName(name),
			Args: pluginConfig[name],
		}
	}
	return ret, nil
}

// scorePluginForSimulator behave like the original plugin.
// But it records the scoring (and normalizedscore) result to store.
type scorePluginForSimulator struct {
	name     string
	original framework.ScorePlugin
	weight   int32

	store *schedulingresultstore.Store
}

const (
	// If original plugin is used for multiple phases(named A and B), we will create plugin for phase A, and the other plugin for phase B.
	// Plugin names should be unique in one scheduler.
	// So, we add phase name as suffix on plugin name.
	scorePluginSuffix = "ForScore"
)

func pluginName(pluginName string) string {
	return pluginName + scorePluginSuffix
}

func NewScoreRecordPlugin(s *schedulingresultstore.Store, p framework.ScorePlugin, weight int32) framework.ScorePlugin {
	sp := &scorePluginForSimulator{
		name:     pluginName(p.Name()),
		original: p,
		weight:   weight,
		store:    s,
	}
	return sp
}

func (pl *scorePluginForSimulator) Name() string { return pl.name }
func (pl *scorePluginForSimulator) ScoreExtensions() framework.ScoreExtensions {
	return pl.original.ScoreExtensions()
}

func (pl *scorePluginForSimulator) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	s := pl.original.ScoreExtensions().NormalizeScore(ctx, state, pod, scores)
	if !s.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", s.Code(), s.Message())
		return s
	}

	for _, s := range scores {
		pl.store.AddNormalizedScoreResult(pod.Namespace, pod.Name, s.Name, pl.original.Name(), s.Score)
	}

	return nil
}

func (pl *scorePluginForSimulator) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	score, s := pl.original.Score(ctx, state, pod, nodeName)
	if !s.IsSuccess() {
		klog.Errorf("failed to run score plugin: %v, %v", s.Code(), s.Message())
		return score, s
	}

	pl.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, pl.original.Name(), score)

	return score, s
}
