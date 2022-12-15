### wasm plugin Part1

It depends on k8s.io/wasm1-proto, and loads k8s.io/wasm1-plugin. (See /staging dir.)

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