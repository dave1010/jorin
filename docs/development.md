# Development and Architecture

This guide covers development workflows, repository layout, and a high-level
architecture overview for jorin.

## Build

From the repository root:

```bash
make build
# or
make
# or
go build -o jorin ./cmd/jorin
```

The Makefile exposes useful targets for development (fmt, test, lint, check).

## Formatting and linting

Run formatting:

```bash
make fmt
make fmt-check
```

Install golangci-lint (recommended) and run lint checks:

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

Integration tests use mock OpenAI servers and local fixtures to validate system prompt composition and tool-call flows end-to-end.

## Make targets

- build: compile binary
- fmt: format code
- fmt-check: check formatting
- lint: run linter
- test: run unit tests
- check: fmt-check + lint + test

## Repository layout

- cmd/jorin: CLI entrypoint and flags
- internal/agent: prompt construction, conversation state, tool orchestration
- internal/openai: OpenAI-compatible client wrapper
- internal/tools: tool implementations and policy checks
- internal/plugins: compiled-in plugin support

## Architecture overview

This project is intentionally small and structured for auditability. Tools are
registered in a registry and invoked via a narrow interface that receives
parameters and returns structured results. Policy checks (readonly, allow/deny,
dry-shell) are evaluated before invoking tools with side effects. This
separation keeps network and filesystem access points concentrated and easy to
review.

## System prompt extensibility

The system prompt sent to the LLM is built from modular prompt providers.
Providers implement a simple interface and can be registered to contribute
parts of the overall system prompt. This makes it easy to add project-specific
instructions, runtime context, or plugin-provided guidance without modifying
core code.

- Providers implement ui.PromptProvider (Provide() string) and are registered
  in init() functions.
- Providers are concatenated in registration order with blank lines between
  sections.
- Default providers include the immutable base instructions, AGENTS.md (when
  present), Skills from ~/.jorin/skills and ./.jorin/skills (SKILL.md
  descriptions), and Situations from ~/.jorin/situations and ./.jorin/situations
  (executables that emit contextual snippets wrapped in XML-like tags).
- The repository ships runtime context Situations in ./.jorin/situations for git
  status, runtime environment, available executables, and Go module detection.

## Contributing

See CONTRIBUTING.md and CODE_OF_CONDUCT.md for contribution guidelines. When
submitting changes, run `make fmt`, `make test`, and `make lint` locally before
opening a PR.
