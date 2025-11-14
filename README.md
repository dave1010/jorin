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
- Examples
- Security model and tool permissions
- Architecture overview
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

The CLI supports a short mode (provide a prompt on the command line), or an
interactive REPL. When invoked with no arguments, `jorin` starts the REPL.

Key flags

- --model: Model ID (default: gpt-5-mini)
- --repl: Start an interactive REPL
- --readonly: Disallow write_file operations
- --dry-shell: Do not execute shell commands (shell calls will be reported as dry run)
- --allow (repeatable): Allowlist substring for shell commands (if set, commands must match at least one allow substring)
- --deny (repeatable): Denylist substring for shell commands (any matching substring will deny execution)
- --cwd: Working directory for tools

## Examples

Start the REPL (also happens if you run the program with no args):

```bash
./jorin --repl
```

Send a single prompt via stdin/pipes (useful in scripts):

```bash
echo "Refactor function X to be smaller" | ./jorin
# or
./jorin "Refactor function X to be smaller"
```

Using allow/deny lists for shell commands:

```bash
# Only allow shell commands that contain the substring ALLOW_ME
./jorin --allow ALLOW_ME "Run the deployment script"

# Deny any shell commands that contain 'rm -rf' or other dangerous substrings
./jorin --deny "rm -rf" "forbidden" "passwd"
```

Prevent the agent from writing files (helpful in auditing runs):

```bash
./jorin --readonly "Make a small change to main.go"
```

Dry-run shell mode (the agent will report what it would run but not execute commands):

```bash
./jorin --dry-shell "Run the tests"
```

## Security model and tool permissions

jorin intentionally exposes a small set of tools to the agent to keep the
attack surface minimal. The agent can request these high-level operations:

- shell: run a shell command (subject to policy allow/deny/dry-run)
- read_file: read a file from disk (requires path)
- write_file: write a file to disk (can be disabled by --readonly)
- http_get: perform an unauthenticated HTTP GET request

Policy controls are provided to limit the agent's capabilities at runtime. The
common controls are:

- Readonly: when enabled (via --readonly), write_file calls are rejected.
- DryShell: when enabled (via --dry-shell), shell calls are not executed; instead
the agent is told what it would have executed.
- Allow list: if one or more --allow values are provided, a shell command will only
run if it contains at least one allow substring.
- Deny list: any shell command containing a deny substring will be rejected.
- CWD: a working directory can be set to confine file operations and shell runs.

Additionally, the CLI appends project-specific instructions when an AGENTS.md
file exists in the working directory. This allows repositories to provide
explicit guidance or constraints to the agent for that project.

When running `jorin` in shared or untrusted environments, prefer using
--readonly and --dry-shell and provide conservative allow/deny lists.

## Architecture overview

The project is organized into a small, explicit pipeline:

- CLI (cmd/jorin): parses flags, builds an initial message set, and manages the REPL.
- Agent (cmd/jorin/agent.go): constructs the system prompt (including project-specific
  instructions and runtime context), maintains conversation state, and coordinates
  chat sessions.
- OpenAI client (internal/openai): handles calls to the OpenAI-compatible API
  (chat once / chat session behavior).
- Tools (internal/tools and registry): a minimal set of tool functions (shell,
  read_file, write_file, http_get) that are invoked via a registry and guarded by
  runtime Policy values.

This small separation makes it straightforward to review and audit the tool
integration points (where external effects occur) and the policy checks that
guard them.

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

## Formatting

Run `gofmt -w .` or use the make target:

```bash
make fmt
```

## Contributing

Contributions are welcome. Please follow these guidelines:

- Read and follow CODE_OF_CONDUCT.md.
- Format and test code. Run `gofmt -w .` and `go test ./...` before opening a PR.
- Keep documentation up to date when you change commands or behavior. Update
  README.md, AGENTS.md, or CHANGELOG.md as appropriate.
- For release-related changes, update CHANGELOG.md (see the template below).

Workflow:

1. Fork the repository and create a branch for your change.
2. Make changes, run `go test ./...`, and ensure formatting is correct.
3. Open a pull request describing your changes and linking any related issues.

If your change is user-facing (CLI behavior, flags, or defaults), include or
update usage examples in README.md.

## CHANGELOG

See CHANGELOG.md for release notes and history.

## License

MIT
