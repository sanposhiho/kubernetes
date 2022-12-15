### wasm plugin

prerequisite: https://github.com/knqyf263/go-plugin#prerequisite

#### build the plugin

```sh
tinygo build -o ./bin/plugin.wasm -scheduler=none -target=wasi --no-debug ./plugin.go
```

But, it won't success because of the limitation of TinyGo.

### About this plugin

It only passes NodeName and Pod object.

```go
import (
	v1 "k8s.io/api/core/v1"
)

type FilterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pod      *v1.Pod `protobuf:"bytes,1,opt,name=pod,proto3" json:"pod,omitempty"`
	NodeName string `protobuf:"bytes,2,opt,name=node_name,json=nodeName,proto3" json:"node_name,omitempty"`
}
```