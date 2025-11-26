# `jorin`

A small coding agent that calls tools (shell, read_file, write_file, http_get)
and communicates with an OpenAI-compatible API. Built to be used as a
command-line tool and REPL.

This README focuses on user-facing instructions. Development and implementation
details have been moved to the docs/ directory. See the docs links below for
more information.

- Quick start and usage: [docs/usage.md](docs/usage.md)
- Development: [docs/development.md](docs/development.md)
- Security and tool permissions: [docs/security.md](docs/security.md)
- Architecture overview: [docs/architecture.md](docs/architecture.md)

## Install

One line install:

```bash
curl -fsSL https://get.jorin.ai | bash
```

Download the latest release for your platform from:

https://github.com/dave1010/jorin/releases

Then add it to your $PATH.

## Quick start

Show help:

```bash
jorin --help
```

Start the REPL (default when invoked with no args):

```bash
jorin
```

Send a single prompt from the command line or a script:

```bash
jorin "Refactor function X to be smaller"
# or
echo "Add logging to foo()" | jorin
```

## Common flags

- --model: Model ID (default: gpt-5-mini)
- --repl: Start an interactive REPL
- --readonly: Disallow write_file operations
- --dry-shell: Do not execute shell commands (shell calls reported as dry run)
- --allow (repeatable): Allowlist substring for shell commands
- --deny (repeatable): Denylist substring for shell commands
- --cwd: Working directory for tools

## Examples

Dry-run shell mode:

```bash
jorin --dry-shell "Run the tests"
```

Prevent file writes:

```bash
jorin --readonly "Make a small change to main.go"
```

For more usage details see docs/.

## Development

### Requirements

- Go toolchain (1.20+ recommended)
- GNU Make (optional)

## Building

Build the CLI:

```bash
make build
# or
go build -o jorin ./cmd/jorin
```

## License

MIT
