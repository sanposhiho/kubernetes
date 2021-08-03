package filter

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func NewRegistryForFilterRecord(s *schedulingresultstore.Store) map[string]schedulerRuntime.PluginFactory {
	ret := map[string]schedulerRuntime.PluginFactory{}
	rs := plugins.NewInTreeRegistry()
	for _, pl := range DefaultFilterPlugins() {
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

	return ret
}

func DefaultFilterPlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	return defaultPlugins.Filter.Enabled
}

// PluginsForSimulator create filterPluginForSimulator for simulator.
// It ignores non-default plugin.
func PluginsForSimulator(disabledPlugins []config.Plugin) []config.Plugin {
	// true means the plugin is disabled
	disabledMap := map[string]bool{}
	for _, p := range disabledPlugins {
		disabledMap[p.Name] = true
	}

	var ret []config.Plugin
	for _, dp := range DefaultFilterPlugins() {
		if !disabledMap[dp.Name] {
			ret = append(ret, config.Plugin{Name: pluginName(dp.Name), Weight: dp.Weight})
		}
	}
	return ret
}

func PluginConfigs() ([]config.PluginConfig, error) {
	defaultPlugins := algorithmprovider.GetDefaultConfig()

	configDecoder := scheme.Codecs.UniversalDecoder()
	ret := make([]config.PluginConfig, len(defaultPlugins.Filter.Enabled))
	for i, p := range defaultPlugins.Filter.Enabled {
		name := p.Name
		gvk := v1beta1.SchemeGroupVersion.WithKind(name + "Args")
		// Use defaults from latest config API version.
		obj, _, err := configDecoder.Decode(nil, &gvk, nil)
		if err != nil {
			if !runtime.IsNotRegisteredError(err) {
				return nil, xerrors.Errorf("get default config: %w", err)
			}
			obj = nil
		}

		ret[i] = config.PluginConfig{
			Name: pluginName(name),
			Args: obj,
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
