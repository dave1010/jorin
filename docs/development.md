# Development

This document covers instructions for building, testing, formatting and
contributing to jorin.

## Build

From the repository root:

```bash
make build
# or
go build -o jorin ./cmd/jorin
```

The Makefile exposes useful targets for development (fmt, test, lint).

## Formatting and linting

Run:

```bash
make fmt
make fmt-check
```

Install golangci-lint (recommended):

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0
make lint
```

## Tests

Run the full test suite:

```bash
make test
# or
go test ./...
```

## Make targets

- build: compile binary
- fmt: format code
- fmt-check: check formatting
- lint: run linter
- test: run unit tests
- check: fmt-check + lint + test

## Code organization

- cmd/jorin: CLI entrypoint and flags
- internal/openai: OpenAI client wrapper
- internal/tools: tool implementations and policy checks
- internal/agent: agent logic and prompt construction

## Contributing

See CONTRIBUTING.md and CODE_OF_CONDUCT.md for contribution guidelines.

When submitting changes, run `make fmt` and `make test` locally before
opening a PR.
