package scorerecord

import (
	"context"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"

	"k8s.io/kubernetes/pkg/scheduler/apis/config"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
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
		ret[scoreRecorderName(pl.Name)] = factory
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
		ret[i] = config.Plugin{Name: scoreRecorderName(n.Name), Weight: n.Weight}
	}
	return ret
}

func ScoreRecordPluginConfig() ([]config.PluginConfig, error) {
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
			Name: scoreRecorderName(name),
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
	scoreRecorderSuffix = "ToRecordScore"
)

func scoreRecorderName(pluginName string) string {
	return pluginName + scoreRecorderSuffix
}

func NewScoreRecordPlugin(s *schedulingresultstore.Store, p framework.ScorePlugin) framework.ScorePlugin {
	return &scoreRecorder{
		name:  scoreRecorderName(p.Name()),
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
	s := pl.p.ScoreExtensions().NormalizeScore(ctx, state, pod, scores)
	if !s.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", s.Code(), s.Message())
		return s
	}

	for _, s := range scores {
		pl.store.AddNormalizeScoreResult(pod.Namespace, pod.Name, s.Name, pl.p.Name(), s.Score)
	}

	return s
}

func (pl *scoreRecorder) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	score, s := pl.p.Score(ctx, state, pod, nodeName)
	if !s.IsSuccess() {
		klog.Errorf("failed to run score plugin: %v, %v", s.Code(), s.Message())
		return score, s
	}

	pl.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, pl.p.Name(), score)

	return score, s
}
