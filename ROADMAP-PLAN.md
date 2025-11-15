# JORIN ROADMAP  REFACTOR PLAN

## Phase 2 — REPL  UI improvements and slash commands

- Add more comprehensive unit/integration tests for StartREPL flows (mock agent to validate forwarding and error handling across more scenarios).
- File-backed history persistence (use only for interactive terminal mode) and history format documentation.
- Multi-line input and paste handling with a configurable terminator (optional; make deterministic for tests).

## Phase 3 — Configuration and CLI subcommands
Goal: Provide consistent configuration precedence and improved CLI UX.

Steps:
1. Add internal/config package to load config from:
   - CLI flags (highest)
   - environment variables
   - project override file (.jorin/config or local file)
   - XDG/standard config location ($XDG_CONFIG_HOME/jorin)
   - defaults
2. Choose a CLI library: cobra + pflag is recommended for subcommands and integration with viper for config. If avoiding viper complexity, implement a small explicit precedence system.
3. Migrate current flag parsing into cobra commands: root command should default to REPL (no args), subcommands: prompt/run/config/sessions/history. Keep short CLI mode (prompt as single arg) as convenience.
4. Wire config into main and pass config struct to agent and tools.

Acceptance criteria:
- CLI exposes clear help and subcommands.
- Config precedence works as documented and is tested with unit tests for loading precedence.

## Phase 4 — Tools, sandboxed shell runner and policy enforcement

Goal: Make shell execution safe and testable, and stabilize tool abstraction.

Steps:
1. Design internal/tools.Tool interface and a registry so that tools are resolved through an interface. Tools receive a context and a structured request object and return structured responses.
2. Implement internal/shell.Runner interface with at least two strategies:
   - Local runner (exec.CommandContext, avoids bash -lc where possible; splits args or shells only when necessary)
   - Dry-run runner (logs but does not execute)
3. Add an optional sandbox wrapper that can be enabled if firejail/nsjail is available. Detect at runtime whether sandbox binary exists and prefer sandbox invocation if configured.
4. Make policy checks explicit and unit-testable: allow list, deny list, dry-shell toggle, readonly toggle. Policy should be in internal/tools or internal/policy and injected into the tool registry.
5. Replace any ad-hoc shell calls with the new shell.Runner implementation.

Acceptance criteria:
- Shell commands executed through a runner with context and timeout.
- Allow/deny/dry-run/readonly policies are fully covered by unit tests.
- No direct calls to exec.CommandContext remain outside the shell package.

## Phase 5 — Sessions, history, and persistence

Goal: Persist sessions for resume/list/delete features.

Steps:
1. Implement internal/session.Store interface with file-backed implementations (JSON, or simple directory of metadata + message files). Use XDG config/ state dirs for storage defaults.
2. Add session metadata (timestamp, model, config snapshot, root working dir). Provide CLI subcommands: sessions list, sessions view id, sessions resume id, sessions delete id.
3. Ensure session I/O is abstracted so tests can use an in-memory store.
4. Document session storage location and format (update README and add examples).

Acceptance criteria:
- sessions subcommands functional and covered by tests for store behavior.
- Session files are stored in configured location with clear metadata layout.

## Phase 6 — File patch tool  SKILLS.md support

Goal: Replace the ad-hoc patching with a more robust tool and support skill manifests.

Steps — file patching:
1. Evaluate existing libraries and approaches (codex-apply-patch, textual patch algorithms). Choose one that does not require the OpenAI Responses API.
2. Implement internal/patch package exposing deterministic: PreparePatch(baseContents, edits) and ApplyPatch(dryRun bool) with revert support.
3. Add tests: dry-run, apply, revert, conflict detection.

Steps — SKILLS.md:
1. Add a parser for a reasonable subset of Claude-style SKILLS.md format and map to the internal tool manifest structure.
2. Provide examples in repo (SKILLS.md sample) and tests that exercise conversion.

Acceptance criteria:
- File patching has test coverage for the main workflows and a safe dry-run mode.
- SKILLS.md parsing can be enabled via config and documented.

## Phase 7 — MCP support and sub-agents

Goal: Add an adapter layer to support MCP and plugin-like model routing; implement sub-agent scaffolding.

Steps:
1. Define openai.Adapter interface that exposes ChatOnce / ChatStream / ModelCapabilities. Implement the current OpenAI-compatible client to satisfy it.
2. Implement an MCP adapter that can be selected via configuration; begin with a stub or mock that demonstrates routing behavior.
3. Design a SubAgent interface (Run(context, request) > result) and a manager that can spawn and supervise sub-agents with scoped permissions and timeouts.
4. Provide examples: a code-gen sub-agent that runs in a restricted workspace, a test-runner sub-agent that executes tests with policy guards.

Acceptance criteria:
- Adapter abstraction exists and is used by the agent.
- Sub-agent interface and a minimal implementation exist with tests.
