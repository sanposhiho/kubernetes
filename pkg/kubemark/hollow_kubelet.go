/*
Copyright 2015 The Kubernetes Authors.

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

package kubemark

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"

	kubeletapp "github.com/sanposhiho/kubernetes/cmd/kubelet/app"
	"github.com/sanposhiho/kubernetes/cmd/kubelet/app/options"
	"github.com/sanposhiho/kubernetes/pkg/kubelet"
	kubeletconfig "github.com/sanposhiho/kubernetes/pkg/kubelet/apis/config"
	"github.com/sanposhiho/kubernetes/pkg/kubelet/cadvisor"
	"github.com/sanposhiho/kubernetes/pkg/kubelet/cm"
	containertest "github.com/sanposhiho/kubernetes/pkg/kubelet/container/testing"
	probetest "github.com/sanposhiho/kubernetes/pkg/kubelet/prober/testing"
	kubetypes "github.com/sanposhiho/kubernetes/pkg/kubelet/types"
	kubeletutil "github.com/sanposhiho/kubernetes/pkg/kubelet/util"
	"github.com/sanposhiho/kubernetes/pkg/util/oom"
	"github.com/sanposhiho/kubernetes/pkg/volume"
	"github.com/sanposhiho/kubernetes/pkg/volume/cephfs"
	"github.com/sanposhiho/kubernetes/pkg/volume/configmap"
	"github.com/sanposhiho/kubernetes/pkg/volume/csi"
	"github.com/sanposhiho/kubernetes/pkg/volume/downwardapi"
	"github.com/sanposhiho/kubernetes/pkg/volume/emptydir"
	"github.com/sanposhiho/kubernetes/pkg/volume/fc"
	"github.com/sanposhiho/kubernetes/pkg/volume/git_repo"
	"github.com/sanposhiho/kubernetes/pkg/volume/hostpath"
	"github.com/sanposhiho/kubernetes/pkg/volume/iscsi"
	"github.com/sanposhiho/kubernetes/pkg/volume/local"
	"github.com/sanposhiho/kubernetes/pkg/volume/nfs"
	"github.com/sanposhiho/kubernetes/pkg/volume/portworx"
	"github.com/sanposhiho/kubernetes/pkg/volume/projected"
	"github.com/sanposhiho/kubernetes/pkg/volume/rbd"
	"github.com/sanposhiho/kubernetes/pkg/volume/secret"
	"github.com/sanposhiho/kubernetes/pkg/volume/util/hostutil"
	"github.com/sanposhiho/kubernetes/pkg/volume/util/subpath"
	"github.com/sanposhiho/kubernetes/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	internalapi "k8s.io/cri-api/pkg/apis"
)

type HollowKubelet struct {
	KubeletFlags         *options.KubeletFlags
	KubeletConfiguration *kubeletconfig.KubeletConfiguration
	KubeletDeps          *kubelet.Dependencies
}

func volumePlugins() []volume.VolumePlugin {
	allPlugins := []volume.VolumePlugin{}
	allPlugins = append(allPlugins, emptydir.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, git_repo.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, hostpath.FakeProbeVolumePlugins(volume.VolumeConfig{})...)
	allPlugins = append(allPlugins, nfs.ProbeVolumePlugins(volume.VolumeConfig{})...)
	allPlugins = append(allPlugins, secret.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, iscsi.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, rbd.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, cephfs.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, downwardapi.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, fc.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, configmap.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, projected.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, portworx.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, local.ProbeVolumePlugins()...)
	allPlugins = append(allPlugins, csi.ProbeVolumePlugins()...)
	return allPlugins
}

func NewHollowKubelet(
	flags *options.KubeletFlags,
	config *kubeletconfig.KubeletConfiguration,
	client *clientset.Clientset,
	heartbeatClient *clientset.Clientset,
	cadvisorInterface cadvisor.Interface,
	imageService internalapi.ImageManagerService,
	runtimeService internalapi.RuntimeService,
	containerManager cm.ContainerManager) *HollowKubelet {
	d := &kubelet.Dependencies{
		KubeClient:               client,
		HeartbeatClient:          heartbeatClient,
		ProbeManager:             probetest.FakeManager{},
		RemoteRuntimeService:     runtimeService,
		RemoteImageService:       imageService,
		CAdvisorInterface:        cadvisorInterface,
		Cloud:                    nil,
		OSInterface:              &containertest.FakeOS{},
		ContainerManager:         containerManager,
		VolumePlugins:            volumePlugins(),
		TLSOptions:               nil,
		OOMAdjuster:              oom.NewFakeOOMAdjuster(),
		Mounter:                  &mount.FakeMounter{},
		Subpather:                &subpath.FakeSubpath{},
		HostUtil:                 hostutil.NewFakeHostUtil(nil),
		PodStartupLatencyTracker: kubeletutil.NewPodStartupLatencyTracker(),
		TracerProvider:           trace.NewNoopTracerProvider(),
		Recorder:                 &record.FakeRecorder{}, // With real recorder we attempt to read /dev/kmsg.
	}

	return &HollowKubelet{
		KubeletFlags:         flags,
		KubeletConfiguration: config,
		KubeletDeps:          d,
	}
}

// Starts this HollowKubelet and blocks.
func (hk *HollowKubelet) Run() {
	if err := kubeletapp.RunKubelet(&options.KubeletServer{
		KubeletFlags:         *hk.KubeletFlags,
		KubeletConfiguration: *hk.KubeletConfiguration,
	}, hk.KubeletDeps, false); err != nil {
		klog.Fatalf("Failed to run HollowKubelet: %v. Exiting.", err)
	}
	select {}
}

// HollowKubeletOptions contains settable parameters for hollow kubelet.
type HollowKubeletOptions struct {
	NodeName            string
	KubeletPort         int
	KubeletReadOnlyPort int
	MaxPods             int
	PodsPerCore         int
	NodeLabels          map[string]string
	RegisterWithTaints  []v1.Taint
}

// Builds a KubeletConfiguration for the HollowKubelet, ensuring that the
// usual defaults are applied for fields we do not override.
func GetHollowKubeletConfig(opt *HollowKubeletOptions) (*options.KubeletFlags, *kubeletconfig.KubeletConfiguration) {
	testRootDir := utils.MakeTempDirOrDie("hollow-kubelet.", "")
	podFilePath := utils.MakeTempDirOrDie("static-pods", testRootDir)
	klog.Infof("Using %s as root dir for hollow-kubelet", testRootDir)

	// Flags struct
	f := options.NewKubeletFlags()
	f.RootDirectory = testRootDir
	f.HostnameOverride = opt.NodeName
	f.MinimumGCAge = metav1.Duration{Duration: 1 * time.Minute}
	f.MaxContainerCount = 100
	f.MaxPerPodContainerCount = 2
	f.NodeLabels = opt.NodeLabels
	f.RegisterSchedulable = true

	// Config struct
	c, err := options.NewKubeletConfiguration()
	if err != nil {
		panic(err)
	}

	c.ImageServiceEndpoint = "unix:///run/containerd/containerd.sock"
	c.StaticPodURL = ""
	c.EnableServer = true
	c.Address = "0.0.0.0" /* bind address */
	c.Port = int32(opt.KubeletPort)
	c.ReadOnlyPort = int32(opt.KubeletReadOnlyPort)
	c.StaticPodPath = podFilePath
	c.FileCheckFrequency.Duration = 20 * time.Second
	c.HTTPCheckFrequency.Duration = 20 * time.Second
	c.NodeStatusUpdateFrequency.Duration = 10 * time.Second
	c.NodeStatusReportFrequency.Duration = 5 * time.Minute
	c.SyncFrequency.Duration = 10 * time.Second
	c.EvictionPressureTransitionPeriod.Duration = 5 * time.Minute
	c.MaxPods = int32(opt.MaxPods)
	c.PodsPerCore = int32(opt.PodsPerCore)
	c.ClusterDNS = []string{}
	c.ImageGCHighThresholdPercent = 90
	c.ImageGCLowThresholdPercent = 80
	c.ProviderID = fmt.Sprintf("kubemark://%v", opt.NodeName)
	c.VolumeStatsAggPeriod.Duration = time.Minute
	c.CgroupRoot = ""
	c.CPUCFSQuota = true
	c.EnableControllerAttachDetach = false
	c.EnableDebuggingHandlers = true
	c.CgroupsPerQOS = false
	// hairpin-veth is used to allow hairpin packets. Note that this deviates from
	// what the "real" kubelet currently does, because there's no way to
	// set promiscuous mode on docker0.
	c.HairpinMode = kubeletconfig.HairpinVeth
	c.MaxOpenFiles = 1024
	c.RegistryBurst = 10
	c.RegistryPullQPS = 5.0
	c.ResolverConfig = kubetypes.ResolvConfDefault
	c.KubeletCgroups = "/kubelet"
	c.SerializeImagePulls = true
	c.SystemCgroups = ""
	c.ProtectKernelDefaults = false
	c.RegisterWithTaints = opt.RegisterWithTaints
	c.RegisterNode = true
	c.LocalStorageCapacityIsolation = true

	return f, c
}
