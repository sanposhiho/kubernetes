package enabledplugin

import (
	"encoding/json"
	"errors"

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
	_, err := GetEnabledPlugin(pod, pluginName, phase)
	if err != nil {
		if !xerrors.Is(err, errPluginNotFound) {
			// something goes wrong
			klog.Errorf("failed to get enabled plugins: %w", err)
		}
		return false
	}

	// plugin is enabled
	return true
}

var (
	errPluginNotFound = errors.New("plugin not found")
)

func GetEnabledPlugin(pod *v1.Pod, pluginName string, phase PluginPhase) (*schedulerapi.Plugin, error) {
	enabledPlugins, err := getEnabledPlugins(pod, phase)
	if err != nil {
		return nil, xerrors.Errorf("get enabled plugin: %w", err)
	}

	for _, p := range enabledPlugins {
		if p.Name == pluginName {
			return &p, nil
		}
	}
	return nil, errPluginNotFound
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

	// TODO: support all plugin phase
	switch phase {
	//	case PreFilter:
	//	case Reserve:
	//	case Permit:
	//	case PreBind:
	//	case Bind:
	//	case PostBind:
	//	case PostFilter:
	//	case PreScore:
	case Filter:
		return plugins.Filter.Enabled, nil
	case Score:
		return plugins.Score.Enabled, nil
	}

	return nil, xerrors.New("non-supported plugin phase")
}
