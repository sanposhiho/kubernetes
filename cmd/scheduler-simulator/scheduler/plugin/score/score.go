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
	manager := newScorePluginManager()
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
			return NewScoreRecordPlugin(s, typed, pl.Weight, manager), nil
		}
		ret[ScorePluginName(pl.Name)] = factory
	}

	return ret
}

func DefaultScorePlugins() []config.Plugin {
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	return defaultPlugins.Score.Enabled
}

func ScorePlugins(enabledPlugins []config.Plugin) []config.Plugin {
	ret := make([]config.Plugin, len(enabledPlugins))
	for i, n := range enabledPlugins {
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

type scorePlugin struct {
	// all scorePlugin should have the same manager
	manager *scorePluginManager

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

func NewScoreRecordPlugin(s *schedulingresultstore.Store, p framework.ScorePlugin, weight int32, manager *scorePluginManager) framework.ScorePlugin {
	sp := &scorePlugin{
		name:          ScorePluginName(p.Name()),
		p:             p,
		defaultWeight: weight,
		store:         s,
		manager:       manager,
	}
	manager.registerPlugin(sp)
	return sp
}

func (pl *scorePlugin) Name() string { return pl.name }
func (pl *scorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (pl *scorePlugin) ApplyPluginWeight(pod *v1.Pod, scores framework.NodeScoreList) {
	weight := enabledplugin.GetPluginWeight(pod, pl.p.Name(), enabledplugin.Score, pl.defaultWeight)
	for i, s := range scores {
		scores[i].Score = s.Score * int64(weight)
	}
}

// NormalizeScore calculates the final score(include plugin weight) of all plugins
// and score 1 for selected node and 0 for the other nodes.
//
// Scheduling framework run score and normalized score in parallel,
// and internally there are upper and lower limits to scores.
// These don't allow score to be changed flexibly from plugins.
// (And, of course, we don't want to change the code in the scheduler itself.)
//
// So, when we run normalized score plugins, it calculates the final score(include plugin weight) of all plugins.
// This allows us to determine which Node will be selected with **user specified** plugin weight.
//
// When run normalized score, we score 1 for selected node and 0 for the other nodes.
func (pl *scorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	selected, s := pl.manager.getSelectedNode(ctx, state, pod)
	if !s.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", s.Code(), s.Message())
		return s
	}
	for _, nodescore := range scores {
		nodescore.Score = 0
		if nodescore.Name == selected {
			// score 1 for selectedNode only.
			nodescore.Score = 1
		}
	}
	return nil
}

func (pl *scorePlugin) RunNormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	if pl.p.ScoreExtensions() != nil {
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
	}

	pl.ApplyPluginWeight(pod, scores)

	for _, s := range scores {
		pl.store.AddFinalScoreResult(pod.Namespace, pod.Name, s.Name, pl.p.Name(), s.Score)
	}

	return nil
}

func (pl *scorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
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

	if err := pl.manager.appendPluginToNodeScores(state, pl.Name(), framework.NodeScore{
		Name:  nodeName,
		Score: score,
	}); err != nil {
		return 0, framework.AsStatus(err)
	}

	return score, s
}
