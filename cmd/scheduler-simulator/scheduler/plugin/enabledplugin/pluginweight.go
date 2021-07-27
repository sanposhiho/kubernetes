package enabledplugin

import (
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// GetPluginWeight gets weight from pod annotation.
// It returns defaultWeight when something goes wrong or weight not found on pod annotation.
func GetPluginWeight(pod *v1.Pod, pluginName string, phase PluginPhase, defaultWeight int32) int32 {
	p, err := GetEnabledPlugin(pod, pluginName, phase)
	if err != nil {
		if !xerrors.Is(err, errPluginNotFound) {
			// something goes wrong
			klog.Errorf("failed to get enabled plugins: %w", err)
		}
		return defaultWeight
	}

	return p.Weight
}
