/*
Copyright 2022 The Kubernetes Authors.

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

package wasm

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	_ "k8s.io/wasm-generated/api/core/v1"
	proto "k8s.io/wasm3-proto"
)

type Wasm struct {
	p proto.WasmScheduling
}

var _ framework.FilterPlugin = &Wasm{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "WASM"
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *Wasm) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *Wasm) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	result, err := pl.p.Filter(ctx, proto.FilterRequest{
		Pod: &proto.IoK8SApiCoreV1Pod{
			Spec: &proto.IoK8SApiCoreV1PodSpec{
				NodeName: pod.Spec.NodeName,
			},
		},
		Node: &proto.IoK8SApiCoreV1Node{
			Metadata: &proto.IoK8SApimachineryPkgApisMetaV1ObjectMeta{
				Name: nodeInfo.Node().Name,
			},
		},
	})
	if err != nil {
		return framework.AsStatus(err)
	}

	if result.StatusCode == int32(framework.Success) {
		return nil
	}

	return framework.NewStatus(framework.Code(result.StatusCode), result.Reason)
}

const pluginPath = "../../../../../staging/src/k8s.io/wasm3-plugin/bin/plugin.wasm"

// New initializes a new plugin and returns it.
func New(_ runtime.Object, f framework.Handle) (framework.Plugin, error) {
	ctx := context.Background()
	p, err := proto.NewWasmSchedulingPlugin(ctx, proto.WasmSchedulingPluginOption{})
	if err != nil {
		return nil, fmt.Errorf("initialize wasmscheduling plugin: %w", err)
	}
	// Ideally we need to close it somewhere.
	// defer p.Close(ctx)

	// Ideally we make pluginPath configurable from the plugin configuration.
	plugin, err := p.Load(ctx, pluginPath)
	if err != nil {
		return nil, fmt.Errorf("load plugin: %w", err)
	}

	return &Wasm{
		p: plugin,
	}, nil
}
