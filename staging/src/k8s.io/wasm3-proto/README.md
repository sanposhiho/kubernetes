### wasm plugin

prerequisite: https://github.com/knqyf263/go-plugin#prerequisite

#### generate SDK from proto

```sh
protoc --go-plugin_out=. --proto_path=../../ --proto_path=. --go-plugin_opt=paths=source_relative ./filter.proto
```

#### generate protobuf for Kubernetes Object 

Use nytimes/openapi2proto, but to make it working with Kubernetes OpenAPI definition, seems we need to use this version: https://github.com/nytimes/openapi2proto/pull/129

```sh
openapi2proto -spec ../../../api/openapi-spec/swagger.json -out kubernetes.proto
```