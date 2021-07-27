package score

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/enabledplugin"
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
			return NewScoreRecordPlugin(s, typed), nil
		}
		ret[ScorePluginName(pl.Name)] = factory
	}

	return ret
}

func DefaultScorePlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	return defaultPlugins.Score.Enabled
}

func ScoreRecorderPlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	ret := make([]config.Plugin, len(defaultPlugins.Score.Enabled))
	for i, n := range defaultPlugins.Score.Enabled {
		ret[i] = config.Plugin{Name: ScorePluginName(n.Name), Weight: n.Weight}
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

type scoreRecorder struct {
	name string
	p    framework.ScorePlugin

	store *schedulingresultstore.Store
}

const (
	scorePluginSuffix = "ForScore"
)

func ScorePluginName(pluginName string) string {
	return pluginName + scorePluginSuffix
}

func NewScoreRecordPlugin(s *schedulingresultstore.Store, p framework.ScorePlugin) framework.ScorePlugin {
	return &scoreRecorder{
		name:  ScorePluginName(p.Name()),
		p:     p,
		store: s,
	}
}

func (pl *scoreRecorder) Name() string { return pl.name }
func (pl *scoreRecorder) ScoreExtensions() framework.ScoreExtensions {
	if pl.p.ScoreExtensions() == nil {
		return nil
	}
	return pl
}

func (pl *scoreRecorder) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	if pl.p.ScoreExtensions() == nil {
		return framework.NewStatus(framework.Error, "this plugin's NormalizeScore should not be called")
	}
	if !enabledplugin.IsPluginEnabled(pod, pl.p.Name(), enabledplugin.Score) {
		for _, s := range scores {
			// When normalizedScore to pass AddFinalScoreResult is -1, it means the plugin is disabled.
			pl.store.AddFinalScoreResult(pod.Namespace, pod.Name, s.Name, pl.p.Name(), schedulingresultstore.DisabledScore)
		}

		return nil
	}

	s := pl.p.ScoreExtensions().NormalizeScore(ctx, state, pod, scores)
	if !s.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", s.Code(), s.Message())
		return s
	}

	for _, s := range scores {
		pl.store.AddFinalScoreResult(pod.Namespace, pod.Name, s.Name, pl.p.Name(), s.Score)
	}

	return s
}

func (pl *scoreRecorder) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	if !enabledplugin.IsPluginEnabled(pod, pl.p.Name(), enabledplugin.Score) {
		pl.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, pl.p.Name(), schedulingresultstore.DisabledScore)

		// return 0 not to affect scoring
		return 0, nil
	}

	score, s := pl.p.Score(ctx, state, pod, nodeName)
	if !s.IsSuccess() {
		klog.Errorf("failed to run score plugin: %v, %v", s.Code(), s.Message())
		return score, s
	}

	pl.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, pl.p.Name(), score)

	return score, s
}
