BINARY := jorin
PKG := ./cmd/jorin
GO := go

# Derive version from latest git tag when building via Makefile. Falls back to "dev".
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)

.PHONY: all build clean fmt fmt-check lint test check

all: build

build:
	$(GO) build -ldflags "-X github.com/dave1010/jorin/internal/version.Version=$(VERSION)" -o $(BINARY) $(PKG)

fmt:
	gofmt -w .

fmt-check:
	@echo "Checking formatting with gofmt..."
	@if [ -n "$(gofmt -l .)" ]; then echo "The following files need formatting:"; gofmt -l .; echo "Run 'gofmt -w .' to fix."; exit 1; fi

lint:
	golangci-lint run --timeout=5m ./...

check: fmt-check lint test

test:
	$(GO) test ./...

clean:
	rm -f $(BINARY)
