.PHONY: build build-all clean test run install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/trabuco/trabuco/internal/cli.Version=$(VERSION)"
INSTALL_DIR := $(HOME)/bin

build:
	go build $(LDFLAGS) -o trabuco ./cmd/trabuco

install: build
	@mkdir -p $(INSTALL_DIR)
	cp trabuco $(INSTALL_DIR)/trabuco
	@echo "Installed trabuco to $(INSTALL_DIR)/trabuco"

build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/trabuco-darwin-amd64 ./cmd/trabuco
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/trabuco-darwin-arm64 ./cmd/trabuco
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/trabuco-linux-amd64 ./cmd/trabuco
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/trabuco-linux-arm64 ./cmd/trabuco
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/trabuco-windows-amd64.exe ./cmd/trabuco

clean:
	rm -rf trabuco dist/

test:
	go test -v ./...

run:
	go run ./cmd/trabuco
