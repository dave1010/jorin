# jorin

A small coding agent that calls tools (shell, read_file, write_file, http_get)
and communicates with an OpenAI-compatible API.
Built to be used as a command-line tool and REPL.

This repository contains the source for the `jorin` command-line agent.

## Table of contents

- Overview
- Requirements
- Build
- Usage
- Makefile targets
- Running tests
- Formatting
- Contributing
- License

## Overview

`jorin` is a lightweight agent that executes tooling (shell commands, file reads/writes, HTTP GET)
and interacts with an OpenAI-compatible API.
The project is implemented in Go and intended to be used from the command line
or invoked programmatically, such as in build scripts or CI actions.

## Requirements

- Go toolchain (1.20+ recommended)
- GNU Make (optional, used for convenience targets)

Ensure `go` is on your PATH before building or running the project.

## Build

From the repository root you can build the CLI binary with:

```bash
go build -o jorin ./cmd/jorin
```

Or use the Makefile target:

```bash
make build
```

## Usage

Show CLI help:

```bash
./jorin --help
```

## Makefile targets

The repository includes a Makefile with common targets to simplify development:

- all (default): builds the agent (same as `make build`).
- build: compiles the binary and writes it to the value of `BINARY` (default: `jorin`).
- fmt: formats Go source files (`gofmt -w .`).
- test: runs the full test suite (`go test ./...`).
- clean: removes the compiled binary (`rm -f $(BINARY)`).

## Running tests

To run the entire test suite from the repository root:

```bash
make test
```

or

```bash
go test ./...
```

## Contributing

Contributions are welcome. Please follow these guidelines:

- Format and test code.
- Keep documentation up to date when you change commands or behavior.

## License

MIT
