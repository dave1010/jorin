# Usage

This document covers typical user-facing ways to run and interact with jorin.

## Quick start

Build the CLI (or download a release):

```bash
make build
# or
go build -o jorin ./cmd/jorin
```

Show help:

```bash
./jorin --help
```

Start the REPL (default when invoked with no args):

```bash
./jorin --repl
# or simply
./jorin
```

Send a single prompt from the command line or a script:

```bash
./jorin "Refactor function X to be smaller"
# or
echo "Add logging to foo()" | ./jorin
```

## Key runtime flags

- --model: Model ID (default: gpt-5-mini)
- --repl: Start an interactive REPL
- --readonly: Disallow write_file operations
- --dry-shell: Do not execute shell commands (shell calls reported as dry run)
- --allow (repeatable): Allowlist substring for shell commands
- --deny (repeatable): Denylist substring for shell commands
- --cwd: Working directory for tools

Use --readonly and --dry-shell together when running in untrusted or shared
environments.

## REPL details

- Interactive terminal (tty): full line editing (cursor movement, history,
  Ctrl-C to abort input).
- Non-interactive (piped): falls back to a simple scanner mode suitable for
  deterministic scripting and tests.
- Slash commands: supported with prefix `/` (e.g. `/help`, `/history`). Escape
  the prefix with a backslash to send a literal slash (e.g. `\/help`).
- Legacy `!` prefix: lines starting with `!` are treated as shell commands.

## Examples

Dry-run shell mode (agent reports shell commands but does not execute them):

```bash
./jorin --dry-shell "Run the tests"
```

Prevent file writes (audit mode):

```bash
./jorin --readonly "Make a small change to main.go"
```

Allow/deny list examples:

```bash
# Only allow shell commands containing the substring ALLOW_ME
./jorin --allow ALLOW_ME "Run the deployment script"

# Deny commands containing dangerous substrings
./jorin --deny "rm -rf" "passwd"
```

For more details about advanced REPL usage, shell policy and tool
permissions see the security and architecture docs:

- [../docs/security.md](docs/security.md)
- [../docs/architecture.md](docs/architecture.md)

If you maintain repository-specific guidance, add an AGENTS.md file to the
project root — jorin will append the contents to the system prompt when run
from that repository.
