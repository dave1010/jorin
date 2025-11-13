BINARY := jorin
PKG := ./cmd/jorin
GO := go

.PHONY: all build clean fmt test

all: build

build:
	$(GO) build -o $(BINARY) $(PKG)

fmt:
	gofmt -w .

test:
	$(GO) test ./...

clean:
	rm -f $(BINARY)
