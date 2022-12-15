### wasm plugin

prerequisite: https://github.com/knqyf263/go-plugin#prerequisite

#### build the plugin

```sh
tinygo build -o ./bin/plugin.wasm -scheduler=none -target=wasi --no-debug ./plugin.go
```

### About this plugin

It only passes NodeName and PodName.

```go
type FilterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PodName  string `protobuf:"bytes,1,opt,name=pod_name,json=podName,proto3" json:"pod_name,omitempty"`
	NodeName string `protobuf:"bytes,2,opt,name=node_name,json=nodeName,proto3" json:"node_name,omitempty"`
}
```