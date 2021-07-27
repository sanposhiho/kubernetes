package schedulerconfig

import (
	// We have to use this config for encoding to json. This is because it has omitempty json tag.
	schedulerapi "k8s.io/kube-scheduler/config/v1beta1"

	// But, we have to use some function which use the type on this package.
	// We call types on this package `originalPlugins` and `originalPluginSet`
	original "k8s.io/kubernetes/pkg/scheduler/apis/config"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
)

func ConvertToOriginalPluginsType(ps *schedulerapi.Plugins) *original.Plugins {
	if ps == nil {
		return nil
	}
	var result original.Plugins
	result.QueueSort = ConvertToOriginalPluginSetType(ps.QueueSort)
	result.PreFilter = ConvertToOriginalPluginSetType(ps.PreFilter)
	result.Filter = ConvertToOriginalPluginSetType(ps.Filter)
	result.PostFilter = ConvertToOriginalPluginSetType(ps.PostFilter)
	result.PreScore = ConvertToOriginalPluginSetType(ps.PreScore)
	result.Score = ConvertToOriginalPluginSetType(ps.Score)
	result.Reserve = ConvertToOriginalPluginSetType(ps.Reserve)
	result.Permit = ConvertToOriginalPluginSetType(ps.Permit)
	result.PreBind = ConvertToOriginalPluginSetType(ps.PreBind)
	result.Bind = ConvertToOriginalPluginSetType(ps.Bind)
	result.PostBind = ConvertToOriginalPluginSetType(ps.PostBind)
	return &result
}

func ConvertToOriginalPluginSetType(originalPluginSet *schedulerapi.PluginSet) original.PluginSet {
	if originalPluginSet == nil {
		return original.PluginSet{}
	}
	convertfn := func(ps []schedulerapi.Plugin) []original.Plugin {
		plugins := make([]original.Plugin, len(ps))
		for i, p := range ps {
			plugins[i].Name = p.Name
			if p.Weight != nil {
				plugins[i].Weight = *p.Weight
			}
		}
		return plugins
	}

	var pluginset original.PluginSet
	pluginset.Enabled = convertfn(originalPluginSet.Enabled)
	pluginset.Disabled = convertfn(originalPluginSet.Disabled)

	return pluginset
}

func ConvertFromOriginalPluginsType(originalPlugins *original.Plugins) *schedulerapi.Plugins {
	var plugins schedulerapi.Plugins
	plugins.QueueSort = ConvertFromOriginalPluginSetType(originalPlugins.QueueSort)
	plugins.PreFilter = ConvertFromOriginalPluginSetType(originalPlugins.PreFilter)
	plugins.Filter = ConvertFromOriginalPluginSetType(originalPlugins.Filter)
	plugins.PostFilter = ConvertFromOriginalPluginSetType(originalPlugins.PostFilter)
	plugins.PreScore = ConvertFromOriginalPluginSetType(originalPlugins.PreScore)
	plugins.Score = ConvertFromOriginalPluginSetType(originalPlugins.Score)
	plugins.Reserve = ConvertFromOriginalPluginSetType(originalPlugins.Reserve)
	plugins.Permit = ConvertFromOriginalPluginSetType(originalPlugins.Permit)
	plugins.PreBind = ConvertFromOriginalPluginSetType(originalPlugins.PreBind)
	plugins.Bind = ConvertFromOriginalPluginSetType(originalPlugins.Bind)
	plugins.PostBind = ConvertFromOriginalPluginSetType(originalPlugins.PostBind)
	return &plugins
}

func ConvertFromOriginalPluginSetType(originalPluginSet original.PluginSet) *schedulerapi.PluginSet {
	convertfn := func(ps []original.Plugin) []schedulerapi.Plugin {
		plugins := make([]schedulerapi.Plugin, len(ps))
		for i, p := range ps {
			p := p
			plugins[i].Name = p.Name
			if p.Weight != 0 {
				plugins[i].Weight = &p.Weight
			}
		}
		return plugins
	}

	var pluginset schedulerapi.PluginSet
	pluginset.Enabled = convertfn(originalPluginSet.Enabled)
	pluginset.Disabled = convertfn(originalPluginSet.Disabled)

	return &pluginset
}

var defaultSchedulerName = scheduler.DefaultSchedulerName

func DefaultSchedulerConfig() *schedulerapi.KubeSchedulerConfiguration {
	ps := algorithmprovider.GetDefaultConfig()
	return &schedulerapi.KubeSchedulerConfiguration{
		TypeMeta: v1.TypeMeta{
			Kind:       "KubeSchedulerConfiguration",
			APIVersion: "kubescheduler.config.k8s.io/v1beta1",
		},
		Profiles: []schedulerapi.KubeSchedulerProfile{
			{
				SchedulerName: &defaultSchedulerName,
				Plugins:       ConvertFromOriginalPluginsType(ps),
			},
		},
	}
}
