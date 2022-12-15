### wasm plugin

prerequisite: https://github.com/knqyf263/go-plugin#prerequisite

#### generate SDK from proto

```sh
protoc --go-plugin_out=. --proto_path=../../ --proto_path=. --go-plugin_opt=paths=source_relative ./filter.proto
```
