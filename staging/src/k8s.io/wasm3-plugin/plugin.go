//go:build tinygo.wasm

package main

import (
	"context"

	proto "k8s.io/wasm3-proto"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	proto.RegisterWasmScheduling(&Plugin{})
}

type Plugin struct{}

var _ proto.WasmScheduling = &Plugin{}

func (p *Plugin) Filter(ctx context.Context, req proto.FilterRequest) (proto.FilterReply, error) {
	pod := req.GetPod()

	if len(pod.Spec.NodeName) == 0 || pod.Spec.NodeName == req.Node.Metadata.Name {
		return proto.FilterReply{
			StatusCode: int32(0), // mark this Node as schedulable. (framwork.Success)
		}, nil
	}

	return proto.FilterReply{
		StatusCode: int32(2), // mark this Node as unschedulable. (framwork.Unschedulable)
	}, nil
}
