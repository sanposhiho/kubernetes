package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	"github.com/sanposhiho/k8s-scheduler-simulator/scheduler/defaultconfig"
)

func Test_convertConfigurationForSimulator(t *testing.T) {
	t.Parallel()

	var nondefaultParallelism int32 = 3
	defaultschedulername := v1.DefaultSchedulerName
	nondefaultschedulername := v1.DefaultSchedulerName + "2"

	type args struct {
		versioned *v1beta2config.KubeSchedulerConfiguration
	}
	tests := []struct {
		name    string
		args    args
		want    *config.KubeSchedulerConfiguration
		wantErr bool
	}{
		{
			name: "success with no-disabled plugin",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta2config.Plugins{},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with empty Profiles",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles does not affects result",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta2config.Plugins{},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles.Plugins does not affects result",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta2config.Plugins{},
							PluginConfig:  nil,
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
						},
						{
							SchedulerName: &nondefaultschedulername,
							Plugins: &v1beta2config.Plugins{
								Score: v1beta2config.PluginSet{
									Disabled: []v1beta2config.Plugin{
										{
											Name: "ImageLocality",
										},
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				profile2 := cfg.Profiles[0].DeepCopy()
				profile2.SchedulerName = nondefaultschedulername
				profile2.Plugins.Score.Enabled = []config.Plugin{
					{Name: "NodeResourcesBalancedAllocationForSimulator", Weight: 1},
					{Name: "InterPodAffinityForSimulator", Weight: 1},
					{Name: "NodeAffinityForSimulator", Weight: 1},
					{Name: "PodTopologySpreadForSimulator", Weight: 2},
					{Name: "TaintTolerationForSimulator", Weight: 1},
				}
				cfg.Profiles = append(cfg.Profiles, *profile2)
				return &cfg
			}(),
		},
		{
			name: "success with some plugin disabled",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &v1beta2config.Plugins{
								Score: v1beta2config.PluginSet{
									Disabled: []v1beta2config.Plugin{
										{
											Name: "ImageLocality",
										},
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.Score.Enabled = []config.Plugin{
					{Name: "NodeResourcesBalancedAllocationForSimulator", Weight: 1},
					{Name: "InterPodAffinityForSimulator", Weight: 1},
					{Name: "NodeAffinityForSimulator", Weight: 1},
					{Name: "PodTopologySpreadForSimulator", Weight: 2},
					{Name: "TaintTolerationForSimulator", Weight: 1},
				}
				return &cfg
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := convertConfigurationForSimulator(tt.args.versioned)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertConfigurationForSimulator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func configGeneratedFromDefault() config.KubeSchedulerConfiguration {
	var weight1 int32 = 1
	var weight2 int32 = 2
	versioned, _ := defaultconfig.DefaultSchedulerConfig()
	cfg := versioned.DeepCopy()
	cfg.Profiles[0].Plugins.Filter.Enabled = []v1beta2config.Plugin{
		{Name: "NodeUnschedulableForSimulator"},
		{Name: "NodeNameForSimulator"},
		{Name: "TaintTolerationForSimulator"},
		{Name: "NodeAffinityForSimulator"},
		{Name: "NodePortsForSimulator"},
		{Name: "NodeResourcesFitForSimulator"},
		{Name: "VolumeRestrictionsForSimulator"},
		{Name: "EBSLimitsForSimulator"},
		{Name: "GCEPDLimitsForSimulator"},
		{Name: "NodeVolumeLimitsForSimulator"},
		{Name: "AzureDiskLimitsForSimulator"},
		{Name: "VolumeBindingForSimulator"},
		{Name: "VolumeZoneForSimulator"},
		{Name: "PodTopologySpreadForSimulator"},
		{Name: "InterPodAffinityForSimulator"},
	}
	cfg.Profiles[0].Plugins.Score.Enabled = []v1beta2config.Plugin{
		{Name: "NodeResourcesBalancedAllocationForSimulator", Weight: &weight1},
		{Name: "ImageLocalityForSimulator", Weight: &weight1},
		{Name: "InterPodAffinityForSimulator", Weight: &weight1},
		{Name: "NodeResourcesFitForSimulator", Weight: &weight1},
		{Name: "NodeAffinityForSimulator", Weight: &weight1},
		{Name: "PodTopologySpreadForSimulator", Weight: &weight2},
		{Name: "TaintTolerationForSimulator", Weight: &weight1},
	}
	pcMap := map[string]runtime.RawExtension{}
	for _, c := range cfg.Profiles[0].PluginConfig {
		pcMap[c.Name] = c.Args
	}

	var newpc []v1beta2config.PluginConfig
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "NodeResourcesBalancedAllocationForSimulator",
		Args: pcMap["NodeResourcesBalancedAllocation"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "InterPodAffinityForSimulator",
		Args: pcMap["InterPodAffinity"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "NodeResourcesFitForSimulator",
		Args: pcMap["NodeResourcesFit"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "NodeAffinityForSimulator",
		Args: pcMap["NodeAffinity"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "PodTopologySpreadForSimulator",
		Args: pcMap["PodTopologySpread"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "VolumeBindingForSimulator",
		Args: pcMap["VolumeBinding"],
	})

	cfg.Profiles[0].PluginConfig = append(cfg.Profiles[0].PluginConfig, newpc...)

	converted := config.KubeSchedulerConfiguration{}
	scheme.Scheme.Convert(cfg, &converted, nil)
	return converted
}
