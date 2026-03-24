BINARY    := bb
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -s -w -X main.version=$(VERSION)

.PHONY: build install test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/bb

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/bb

test:
	go test ./... -race -cover

lint:
	golangci-lint run

clean:
	rm -rf bin/
