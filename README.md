# agent

A small coding agent that calls tools (shell, read_file, write_file, http_get) and speaks to an OpenAI-compatible API. Built to be used as a command-line tool and REPL.

Usage:
  go build -o agent ./cmd/agent
  ./agent --help

The module path is github.com/dave1010/agent

Makefile targets

This repository includes a Makefile with the following targets to simplify common tasks:

- all (default): builds the agent (same as `make build`).
- build: compiles the binary and writes it to the value of BINARY (default: `agent`). Use `make build` or `make`.
- fmt: runs `gofmt -w .` to format all Go files in the repository.
- test: runs `go test ./...` to execute the full test suite.
- clean: removes the compiled binary (runs `rm -f $(BINARY)`).

Examples:

  # build the agent
  make build

  # format all Go files
  make fmt

  # run the full test suite
  make test

  # clean the compiled binary
  make clean

Running tests

This repository includes unit tests for the cmd/agent package. To run the full test suite from the repository root:

  go test ./...

For more verbose output:

  go test ./... -v

Run tests for the agent package only:

  go test ./cmd/agent -v

Run a single test by name (replace TestName with the test function):

  go test ./cmd/agent -run TestName -v

Formatting and linting

Before committing changes to Go code, format files with:

  gofmt -w .

This project follows standard Go tooling; ensure you have Go installed and available on your PATH.
