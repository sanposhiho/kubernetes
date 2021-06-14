.PHONY: tools
tools:
	cd ./tools; \
	cat tools.go | grep "_" | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: format
format:
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test ./...
