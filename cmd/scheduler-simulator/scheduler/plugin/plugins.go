package plugin

import (
	"golang.org/x/xerrors"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/filter"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/score"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func NewRegistry(informerFactory informers.SharedInformerFactory, client clientset.Interface) schedulerRuntime.Registry {
	defaultScorePluginWeight := map[string]int32{}
	defaultScorePlugin := score.DefaultScorePlugins()
	for _, p := range defaultScorePlugin {
		defaultScorePluginWeight[p.Name] = p.Weight
	}

	store := schedulingresultstore.New(informerFactory, client, defaultScorePluginWeight)
	sr := score.NewRegistryForScoreRecord(store)
	fr := filter.NewRegistryForFilterRecord(store)

	return merge(sr, fr)
}

// merge merges multiple map.
func merge(pfs ...map[string]schedulerRuntime.PluginFactory) map[string]schedulerRuntime.PluginFactory {
	ret := map[string]schedulerRuntime.PluginFactory{}
	for _, pf := range pfs {
		for n, r := range pf {
			ret[n] = r
		}
	}
	return ret
}

func NewPlugins() *config.Plugins {
	defaultFilterPlugins := filter.DefaultFilterPlugins()
	defaultScorePlugins := score.DefaultScorePlugins()
	return &config.Plugins{
		Filter: config.PluginSet{
			// enable all filter plugins for simulator
			Enabled: filter.FilterPlugins(defaultFilterPlugins),
			// disable all default filter plugin.
			Disabled: defaultFilterPlugins,
		},
		Score: config.PluginSet{
			// enable all score plugins for simulator
			Enabled: score.ScorePlugins(defaultScorePlugins),
			// disable all default score plugin.
			Disabled: defaultScorePlugins,
		},
	}
}

func NewPluginConfig() ([]config.PluginConfig, error) {
	s, err := score.PluginConfigs()
	if err != nil {
		return nil, xerrors.Errorf("get score record plugin config: %w", err)
	}
	f, err := filter.PluginConfigs()
	if err != nil {
		return nil, xerrors.Errorf("get filter record plugin config: %w", err)
	}
	return append(s, f...), nil
}
