package util

import (
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/output/scheme"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
)

// DefaultSchedulerConfig creates KubeSchedulerConfiguration default configuration.
func DefaultSchedulerConfig() (*config.KubeSchedulerConfiguration, error) {
	var versionedCfg v1beta2.KubeSchedulerConfiguration
	scheme.Scheme.Default(&versionedCfg)
	cfg := config.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(&versionedCfg, &cfg, nil); err != nil {
		return nil, err
	}

	return &cfg, nil
}
