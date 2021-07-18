package enabledpluginutil

import (
	"encoding/json"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	// enable plugins are recorded on annotation with enabledPluginsAnnotationKey.
	enabledPluginsAnnotationKey = "scheduler-simulator/enabled-plugins"
)

func IsPluginEnabled(pod *v1.Pod, pluginName string) bool {
	enabledPlugins, err := getEnabledPlugins(pod)
	if err != nil || enabledPlugins == nil {
		if err != nil {
			klog.Errorf("failed to get enabled plugins: %w", err)
		}
		// when enabledPluginAnnotation does not exist or something goes wrong with get enabled plugins,
		// all plugins will be enabled.
		return false
	}

	for _, p := range enabledPlugins {
		if p == pluginName {
			return true
		}
	}

	return false
}

// getEnabledPlugins get enabled plugins from pod annotation with enabledPluginsAnnotationKey.
func getEnabledPlugins(pod *v1.Pod) ([]string, error) {
	j, ok := pod.Annotations[enabledPluginsAnnotationKey]
	if !ok {
		return nil, xerrors.Errorf("%s is not found on pod annotation", enabledPluginsAnnotationKey)
	}

	var enabled []string
	if err := json.Unmarshal([]byte(j), &enabled); err != nil {
		return nil, xerrors.Errorf("encode json: %w", err)
	}

	return enabled, nil
}
