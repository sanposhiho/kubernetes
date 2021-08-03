package score

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func NewRegistryForScoreRecord(s *schedulingresultstore.Store) map[string]schedulerRuntime.PluginFactory {
	ret := map[string]schedulerRuntime.PluginFactory{}
	rs := plugins.NewInTreeRegistry()
	for _, pl := range DefaultScorePlugins() {
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
		ret[ScorePluginName(pl.Name)] = factory
	}

	return ret
}

func DefaultScorePlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	return defaultPlugins.Score.Enabled
}

// ScorePlugins create scorePlugin for simulator.
func ScorePlugins(disabledPlugins []config.Plugin) []config.Plugin {
	// true means the plugin is disabled
	disabledMap := map[string]bool{}
	for _, p := range disabledPlugins {
		disabledMap[p.Name] = true
	}

	var ret []config.Plugin
	for _, dp := range DefaultScorePlugins() {
		if !disabledMap[dp.Name] {
			ret = append(ret, config.Plugin{Name: ScorePluginName(dp.Name), Weight: dp.Weight})
		}
	}
	return ret
}

func PluginConfigs() ([]config.PluginConfig, error) {
	defaultPlugins := algorithmprovider.GetDefaultConfig()

	configDecoder := scheme.Codecs.UniversalDecoder()
	ret := make([]config.PluginConfig, len(defaultPlugins.Score.Enabled))
	for i, p := range defaultPlugins.Score.Enabled {
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
			Name: ScorePluginName(name),
			Args: obj,
		}
	}
	return ret, nil
}

// TODO: write doc about this.
type scorePlugin struct {
	name          string
	p             framework.ScorePlugin
	defaultWeight int32

	store *schedulingresultstore.Store
}

const (
	scorePluginSuffix = "ForScore"
)

func ScorePluginName(pluginName string) string {
	return pluginName + scorePluginSuffix
}

func NewScoreRecordPlugin(s *schedulingresultstore.Store, p framework.ScorePlugin, weight int32) framework.ScorePlugin {
	sp := &scorePlugin{
		name:          ScorePluginName(p.Name()),
		p:             p,
		defaultWeight: weight,
		store:         s,
	}
	return sp
}

func (pl *scorePlugin) Name() string { return pl.name }
func (pl *scorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl.p.ScoreExtensions()
}

func (pl *scorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	s := pl.p.ScoreExtensions().NormalizeScore(ctx, state, pod, scores)
	if !s.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", s.Code(), s.Message())
		return s
	}

	for _, s := range scores {
		pl.store.AddNormalizedScoreResult(pod.Namespace, pod.Name, s.Name, pl.p.Name(), s.Score)
	}

	return nil
}

func (pl *scorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	score, s := pl.p.Score(ctx, state, pod, nodeName)
	if !s.IsSuccess() {
		klog.Errorf("failed to run score plugin: %v, %v", s.Code(), s.Message())
		return score, s
	}

	pl.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, pl.p.Name(), score)

	return score, s
}
