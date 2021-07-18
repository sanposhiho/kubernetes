package filterrecord

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugins/enabledpluginutil"
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
		ret[filterRecorderName(pl.Name)] = factory
	}

	return ret
}

func DefaultFilterPlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	return defaultPlugins.Filter.Enabled
}

func FilterRecorderPlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	ret := make([]config.Plugin, len(defaultPlugins.Filter.Enabled))
	for i, n := range defaultPlugins.Filter.Enabled {
		ret[i] = config.Plugin{Name: filterRecorderName(n.Name), Weight: n.Weight}
	}
	return ret
}

func FilterRecordPluginConfig() ([]config.PluginConfig, error) {
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
			Name: filterRecorderName(name),
			Args: obj,
		}
	}
	return ret, nil
}

type filterRecorder struct {
	name string
	p    framework.FilterPlugin

	store *schedulingresultstore.Store
}

const (
	filterRecorderSuffix = "ToRecordFilteringResult"
)

func filterRecorderName(pluginName string) string {
	return pluginName + filterRecorderSuffix
}

func NewFilterRecordPlugin(s *schedulingresultstore.Store, p framework.FilterPlugin) framework.FilterPlugin {
	return &filterRecorder{
		name:  filterRecorderName(p.Name()),
		p:     p,
		store: s,
	}
}

func (pl *filterRecorder) Name() string { return pl.name }

func (pl *filterRecorder) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if !enabledpluginutil.IsPluginEnabled(pod, pl.p.Name()) {
		pl.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, pl.p.Name(), "(disabled)")
		// return success not to affect filtering.
		return nil
	}

	s := pl.p.Filter(ctx, state, pod, nodeInfo)
	if s.IsSuccess() {
		if s == nil {
			// When status is nil (= success), s.AppendReason panic.
			s = framework.NewStatus(framework.Success)
		}
		s.AppendReason("passed")
	}

	pl.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, pl.p.Name(), s.Message())
	return s
}
