package filter

// TODO merge this package with score package on ../score/score.go

import (
	"context"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/util"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func NewRegistryForFilterRecord(s *schedulingresultstore.Store) (map[string]schedulerRuntime.PluginFactory, error) {
	ret := map[string]schedulerRuntime.PluginFactory{}
	rs := plugins.NewInTreeRegistry()

	defaultpls, err := DefaultFilterPlugins()
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
			typed, ok := p.(framework.FilterPlugin)
			if !ok {
				return nil, xerrors.New("not filter plugin")
			}
			return NewFilterRecordPlugin(s, typed), nil
		}
		ret[pluginName(pl.Name)] = factory
	}

	return ret, nil
}

func DefaultFilterPlugins() ([]config.Plugin, error) {
	defaultConfig, err := util.DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return nil, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Filter.Enabled, nil
}

// PluginsForSimulator create filterPluginForSimulator for simulator.
// It ignores non-default plugin.
func PluginsForSimulator(disabledPlugins []config.Plugin) ([]config.Plugin, error) {
	// true means the plugin is disabled
	disabledMap := map[string]bool{}
	for _, p := range disabledPlugins {
		disabledMap[p.Name] = true
	}

	defaultpls, err := DefaultFilterPlugins()
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
	defaultpls, err := DefaultFilterPlugins()
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

// filterPluginForSimulator behave like the original plugin.
// But it records the filtering result to store.
type filterPluginForSimulator struct {
	name     string
	original framework.FilterPlugin

	store *schedulingresultstore.Store
}

const (
	// If original plugin is used for multiple phases(named A and B), we will create plugin for phase A, and the other plugin for phase B.
	// Plugin names should be unique in one scheduler.
	// So, we add phase name as suffix on plugin name.
	filterPluginSuffix = "ForFilter"
)

func pluginName(pluginName string) string {
	return pluginName + filterPluginSuffix
}

func NewFilterRecordPlugin(s *schedulingresultstore.Store, p framework.FilterPlugin) framework.FilterPlugin {
	return &filterPluginForSimulator{
		name:     pluginName(p.Name()),
		original: p,
		store:    s,
	}
}

func (pl *filterPluginForSimulator) Name() string { return pl.name }

func (pl *filterPluginForSimulator) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	s := pl.original.Filter(ctx, state, pod, nodeInfo)
	if s.IsSuccess() {
		pl.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, pl.original.Name(), schedulingresultstore.PassedFilterMessage)
		return s
	}

	pl.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, pl.original.Name(), s.Message())
	return s
}
