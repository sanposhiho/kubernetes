//go:build tinygo.wasm

package main

import (
	"context"

	proto "k8s.io/wasm1-proto"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	proto.RegisterWasmScheduling(&Plugin{})
}

type Plugin struct{}

var _ proto.WasmScheduling = &Plugin{}

func (p *Plugin) Filter(ctx context.Context, req proto.FilterRequest) (proto.FilterReply, error) {
	if req.NodeName == "good-node" {
		return proto.FilterReply{
			StatusCode: int32(0), // Pass this Node. (framework.Success)
		}, nil
	}

	return proto.FilterReply{
		StatusCode: int32(2), // mark this Node as unschedulable. (framwork.Unschedulable)
	}, nil
}
