package enabledplugin

import (
	"encoding/json"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/annotation"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/config"
)

type PluginPhase int

const (
	PreFilter PluginPhase = iota
	Filter
	PostFilter
	PreScore
	Score
	Reserve
	Permit
	PreBind
	Bind
	PostBind
)

func IsPluginEnabled(pod *v1.Pod, pluginName string, phase PluginPhase) bool {
	enabledPlugins, err := getEnabledPlugins(pod, phase)
	if err != nil || enabledPlugins == nil {
		if err != nil {
			klog.Errorf("failed to get enabled plugins: %w", err)
		}
		// when enabledPluginAnnotation does not exist or something goes wrong with get enabled plugins,
		// all plugins will be enabled.
		return false
	}

	for _, p := range enabledPlugins {
		if p.Name == pluginName {
			return true
		}
	}

	return false
}

// getEnabledPlugins get enabled plugins from pod annotation with enabledPluginsAnnotationKey.
func getEnabledPlugins(pod *v1.Pod, phase PluginPhase) ([]schedulerapi.Plugin, error) {
	j, ok := pod.Annotations[annotation.EnabledPluginsAnnotationKey]
	if !ok {
		return nil, xerrors.Errorf("%s is not found on pod annotation", annotation.EnabledPluginsAnnotationKey)
	}

	var plugins schedulerapi.Plugins
	if err := json.Unmarshal([]byte(j), &plugins); err != nil {
		return nil, xerrors.Errorf("encode plugin json: %w", err)
	}

	switch phase {
	//	case PreFilter:
	case Filter:
		return plugins.Filter.Enabled, nil
		//	case PostFilter:
		//	case PreScore:
	case Score:
		return plugins.Score.Enabled, nil
		//	case Reserve:
		//	case Permit:
		//	case PreBind:
		//	case Bind:
		//	case PostBind:
	}

	// TODO: support all plugins
	return nil, xerrors.New("non-supported plugin phase")
}
