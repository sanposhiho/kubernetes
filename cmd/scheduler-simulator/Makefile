.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: format
format:
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test ./...
