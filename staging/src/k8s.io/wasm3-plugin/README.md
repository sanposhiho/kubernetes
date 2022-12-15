### wasm plugin

prerequisite: https://github.com/knqyf263/go-plugin#prerequisite

#### generate SDK from proto

```sh
protoc --go-plugin_out=. --proto_path=../../../../../staging/src  --proto_path=. --go-plugin_opt=paths=source_relative ./proto/filter.proto
```

#### generate golang types

prerequisite: https://github.com/kubewarden/k8s-objects-generator#requirements

```sh
k8s-objects-generator -f ../../../../../api/openapi-spec/swagger.json -o . -repo k8s.io/kubernetes/pkg/scheduler/framework/plugins/wasm3/generated
find src/k8s.io/kubernetes/pkg/scheduler/framework/plugins/wasm3/generated/* -maxdepth 0 | xargs -I% mv % ./generated/ 

# in the k/k root.
```

#### build the plugin

```sh
tinygo build -o ./bin/plugin.wasm -scheduler=none -target=wasi --no-debug ./plugin.go
```