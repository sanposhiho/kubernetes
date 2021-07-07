.PHONY: tools
tools:
	cd ./tools; \
	cat tools.go | grep "_" | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: generate
generate:
	go generate ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: format
format:
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o ./bin/simulator ./simulator.go

.PHONY: serve
serve: build
	./hack/start_simulator.sh