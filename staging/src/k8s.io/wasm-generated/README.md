#### generate golang types compilable with TinyGo

prerequisite: https://github.com/kubewarden/k8s-objects-generator#requirements

```sh
cd ../ 
# in /staging/src/k8s.io dir.
k8s-objects-generator -f ./../../../../api/openapi-spec/swagger.json -o ./wasm-generated -repo k8s.io/wasm-generated
mv ./wasm-generated/src/k8s.io/wasm-generated/* ./wasm-generated && rm -rf src
```

#### generate protobuf from swagger

prerequisite: https://github.com/nytimes/openapi2proto

Example: 

```sh
openapi2proto -spec  api/core/v1/swagger.json -out api/core/v1/v1.proto

# add `option go_package = "k8s.io/wasm3-generated/api/core/v1";` to api/core/v1/v1.proto.
```