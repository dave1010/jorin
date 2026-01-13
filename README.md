# `jorin`

A small coding agent that calls tools (shell, read_file, write_file, http_get)
and communicates with an OpenAI-compatible API. Built to be used as a
command-line tool and REPL.

This README focuses on user-facing instructions. Development and implementation
details have been moved to the docs/ directory. See the docs links below for
more information.

- Quick start and usage: [Usage guide](docs/usage.md)
- CLI reference: [CLI reference](docs/reference.md)
- Development: [Development guide](docs/development.md)
- Security and tool permissions: [Security notes](docs/security.md)
- Architecture overview: [Architecture overview](docs/architecture.md)
- Troubleshooting: [Troubleshooting](docs/troubleshooting.md)

## Install

One line install:

```bash
curl -fsSL https://get.jorin.ai | bash
```

Download the latest release for your platform from
[GitHub Releases](https://github.com/dave1010/jorin/releases).

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

For more usage details, see the [usage guide](docs/usage.md).

## Skills and Situations

Jorin can load optional prompt context from two conventions:

- **Skills** (Anthropic convention): `~/.jorin/skills` or `./.jorin/skills` with a
  `SKILL.md` file that contains YAML frontmatter (`name`, `description`). Jorin
  injects the descriptions into the system prompt and instructs the agent to
  read the full skill file when relevant. See
  <https://code.claude.com/docs/en/skills> for the Skills convention.
- **Situations** (Jorin-specific): `~/.jorin/situations` or `./.jorin/situations`
  directories containing a `SITUATION.yaml` file and an executable referenced
  by its `run` field. The executable output is wrapped in `<name>...</name>`
  tags and appended to the system prompt. The repo ships built-in situations
  under `./.jorin/situations` (env, execs, git, go). See
  <https://github.com/dave1010/agent-situations> and
  <https://dave.engineer/blog/2026/01/agent-situations/> for information on
  Situations.

For more usage details and setup steps, see the
[usage guide](docs/usage.md).

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
