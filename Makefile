BINARY    := bb
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -s -w -X main.version=$(VERSION)

.PHONY: build install test lint clean tidy

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/bb

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/bb

test:
	go test ./... -race -cover

lint:
	golangci-lint run

# Use GOTOOLCHAIN=local so go mod tidy never bumps the go directive in go.mod
tidy:
	go mod tidy

clean:
	rm -rf bin/
