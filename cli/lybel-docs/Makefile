# lybel-docs Makefile
#
# Common targets:
#   make build       Build native binary into bin/lybel-docs
#   make build-all   Cross-compile for darwin/linux/windows × amd64/arm64
#   make test        Run unit tests
#   make install     Build and copy into ../../skills/lybel-docs/bin/lybel-docs
#   make clean       Remove bin/ and dist/
#   make fmt         go fmt ./...
#   make lint        go vet ./...

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
PKG     := .
BIN     := lybel-docs

INSTALL_DIR := ../../skills/lybel-docs/bin

.PHONY: all build build-all test install clean fmt lint deps

all: build

deps:
	go mod tidy

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BIN) $(PKG)

build-all: clean-dist
	@mkdir -p dist
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-darwin-amd64    $(PKG)
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-darwin-arm64    $(PKG)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-linux-amd64     $(PKG)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-linux-arm64     $(PKG)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-windows-amd64.exe $(PKG)
	@echo "built dist/ artifacts for version $(VERSION)"

test:
	go test -v ./...

install: build
	@mkdir -p $(INSTALL_DIR)
	cp bin/$(BIN) $(INSTALL_DIR)/$(BIN)
	@echo "installed to $(INSTALL_DIR)/$(BIN)"

clean: clean-dist
	rm -rf bin

clean-dist:
	rm -rf dist

fmt:
	go fmt ./...

lint:
	go vet ./...
