/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

// This file exists to force the desired plugin implementations to be linked.
import (
	"k8s.io/component-base/featuregate"
	"k8s.io/utils/exec"

	// Volume plugins
	"github.com/sanposhiho/kubernetes/pkg/volume"
	"github.com/sanposhiho/kubernetes/pkg/volume/cephfs"
	"github.com/sanposhiho/kubernetes/pkg/volume/configmap"
	"github.com/sanposhiho/kubernetes/pkg/volume/csi"
	"github.com/sanposhiho/kubernetes/pkg/volume/downwardapi"
	"github.com/sanposhiho/kubernetes/pkg/volume/emptydir"
	"github.com/sanposhiho/kubernetes/pkg/volume/fc"
	"github.com/sanposhiho/kubernetes/pkg/volume/flexvolume"
	"github.com/sanposhiho/kubernetes/pkg/volume/git_repo"
	"github.com/sanposhiho/kubernetes/pkg/volume/hostpath"
	"github.com/sanposhiho/kubernetes/pkg/volume/iscsi"
	"github.com/sanposhiho/kubernetes/pkg/volume/local"
	"github.com/sanposhiho/kubernetes/pkg/volume/nfs"
	"github.com/sanposhiho/kubernetes/pkg/volume/projected"
	"github.com/sanposhiho/kubernetes/pkg/volume/secret"

	// Cloud providers
	_ "github.com/sanposhiho/kubernetes/pkg/cloudprovider/providers"
)

// ProbeVolumePlugins collects all volume plugins into an easy to use list.
func ProbeVolumePlugins(featureGate featuregate.FeatureGate) ([]volume.VolumePlugin, error) {
	allPlugins := []volume.VolumePlugin{}

	// The list of plugins to probe is decided by the kubelet binary, not
	// by dynamic linking or other "magic".  Plugins will be analyzed and
	// initialized later.
	//
	// Kubelet does not currently need to configure volume plugins.
	// If/when it does, see kube-controller-manager/app/plugins.go for example of using volume.VolumeConfig
	var err error
	allPlugins, err = appendLegacyProviderVolumes(allPlugins, featureGate)
	if err != nil {
		return allPlugins, err
	}
	allPlugins = append(allPlugins, emptydir.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, git_repo.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, hostpath.ProbeVolumePlugins(volume.VolumeConfig{})...)
	allPlugins = append(allPlugins, nfs.ProbeVolumePlugins(volume.VolumeConfig{})...)
	allPlugins = append(allPlugins, secret.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, iscsi.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, cephfs.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, downwardapi.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, fc.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, configmap.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, projected.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, local.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, csi.ProbeVolumePlugins()...)
	return allPlugins, nil
}

// GetDynamicPluginProber gets the probers of dynamically discoverable plugins
// for kubelet.
// Currently only Flexvolume plugins are dynamically discoverable.
func GetDynamicPluginProber(pluginDir string, runner exec.Interface) volume.DynamicPluginProber {
	return flexvolume.GetDynamicPluginProber(pluginDir, runner)
}
