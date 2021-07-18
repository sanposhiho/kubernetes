package plugins

import (
	"golang.org/x/xerrors"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugins/filterrecord"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugins/scorerecord"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/schedulingresultstore"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func NewRegistry(informerFactory informers.SharedInformerFactory, client clientset.Interface) schedulerRuntime.Registry {
	store := schedulingresultstore.New(informerFactory, client)
	sr := scorerecord.NewRegistryForScoreRecord(store)
	fr := filterrecord.NewRegistryForFilterRecord(store)

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

func NewPlugin() *config.Plugins {
	return &config.Plugins{
		Filter: config.PluginSet{
			Enabled: filterrecord.FilterRecorderPlugins(),
			// disable all filter plugin.
			Disabled: filterrecord.DefaultFilterPlugins(),
		},
		Score: config.PluginSet{
			Enabled: scorerecord.ScoreRecorderPlugins(),
			// disable all score plugin.
			Disabled: scorerecord.DefaultScorePlugins(),
		},
	}
}

func NewPluginConfig() ([]config.PluginConfig, error) {
	s, err := scorerecord.ScoreRecordPluginConfig()
	if err != nil {
		return nil, xerrors.Errorf("get score record plugin config: %w", err)
	}
	f, err := filterrecord.FilterRecordPluginConfig()
	if err != nil {
		return nil, xerrors.Errorf("get filter record plugin config: %w", err)
	}
	return append(s, f...), nil
}
