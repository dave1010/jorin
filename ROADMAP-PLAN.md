JORIN ROADMAP  REFACTOR PLAN

Purpose

This document outlines an ordered plan to refactor and stabilize the jorin codebase so it is maintainable, testable, and ready for the feature work described in the open issues. The emphasis for the initial phase is ensuring the codebase is in a good state (clear package boundaries, DI-friendly, well-tested, documented), then following with targeted, concrete feature implementations.

Summary of relevant GitHub issues (high level)

- #1 Improve the text UI (TUI, testability of REPL)
- #2 Sub agents (agent delegation / sub-agents)
- #3 Slash commands in REPL
- #4 Configuration (XDG locations, precedence, viper/cobra suggestions)
- #5 Support for Claude-style SKILLS.md
- #6 MCP support (model control protocol / adapter architecture)
- #7 Sandbox shell (firejail/nsjail, avoid bash -lc, use exec.CommandContext)
- #8 Better file patch tool (codex-apply-patch, robust patch workflows)
- #9 Chat history and sessions (persist/load/list sessions)
- #10 Sub commands (clearer subcommand structure, use cobra/pflag)
- #11 Refactor: split responsibilities into packages

High-level objectives (what "good state" looks like)

- Clear package layout: cmd/, internal/{agent,openai,tools,ui,config,session,...}
- Well-defined interfaces for external-effecting components (tools, shell runner, file ops, HTTP client, model adapter)
- Minimal global state; enable dependency injection so the CLI can wire components for production and tests
- startREPL accepts io.Reader/io.Writer for unit tests and embedding
- Tool registry and policy checks are interface-driven and unit-tested
- Thorough unit tests for non-UI logic; small integration tests for CLI behaviors where possible
- Formatting and static checks enforced (gofmt, go vet, golangci-lint)
- Documentation updated for config, expected runtime behavior, and developer workflow

Priority order (what to do first and why)

1) Split responsibilities into packages and define clear interfaces (#11)
   - Rationale: foundational refactor; enables all other work (CLI changes, sessions, MCP, sandboxing). Make this the first, small-step priority.
2) Make REPL and UI testable (#1, #3)
   - Rationale: many features (slash commands, history) affect REPL; testability avoids regressions.
3) Introduce a configuration layer and CLI subcommand framework (#4, #10)
   - Rationale: many features require config and flags; using a standard framework (cobra + viper or pflag) will make future flags consistent.
4) Stabilize the tools abstraction and sandboxed shell runner (#7)
   - Rationale: safety and auditability; needed for file write policy, dry-run behavior, allow/deny checks.
5) Implement sessions / chat history persistence (#9)
   - Rationale: user-facing persistence; depends on agent interfaces and config paths.
6) Add a better file patch tool and structured tool manifest (SKILLS.md) (#8, #5)
   - Rationale: higher-value features that rely on a solid tools/IO design.
7) MCP support and sub-agents (#6, #2)
   - Rationale: larger features that build on an extensible adapter architecture.

Detailed plan of attack (concrete steps)

Phase 1 — Package refactor and interfaces (foundation)
COMPLETE

Goal: move towards a modular layout without large behavior changes. Small, review-friendly commits.

Phase 2 — REPL  UI improvements and slash commands

Overview

Phase 2 focuses on extracting UI and REPL logic into small, well-tested packages and on adding a small, deterministic slash-command parser  dispatcher. Changes should be behavior-preserving for the default interactive experience while enabling automated tests and future TUI enhancements.

Remaining tasks for Phase 2

- Add more comprehensive unit/integration tests for StartREPL flows (mock agent to validate forwarding and error handling across more scenarios).
- Implement optional terminal adapter and line-editing support for interactive sessions (currently StartREPL uses bufio.Scanner; consider optional build-tag gated support for advanced line-editing later).
- File-backed history persistence (use only for interactive terminal mode) and history format documentation.
- Multi-line input and paste handling with a configurable terminator (optional; make deterministic for tests).
- Documentation updates in README.md about slash commands and escaping conventions.

Notes on scope

The intention here is to keep Phase 2 focused and incremental: the codebase already contains the core parser, handler, StartREPL wiring, and a simple history implementation. Remaining work should focus on test coverage, persistence, and optional UX improvements.

Phase 3 — Configuration and CLI subcommands
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

Phase 4 — Tools, sandboxed shell runner and policy enforcement
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

Phase 5 — Sessions, history, and persistence
Goal: Persist sessions for resume/list/delete features.

Steps:
1. Implement internal/session.Store interface with file-backed implementations (JSON, or simple directory of metadata + message files). Use XDG config/ state dirs for storage defaults.
2. Add session metadata (timestamp, model, config snapshot, root working dir). Provide CLI subcommands: sessions list, sessions view id, sessions resume id, sessions delete id.
3. Ensure session I/O is abstracted so tests can use an in-memory store.
4. Document session storage location and format (update README and add examples).

Acceptance criteria:
- sessions subcommands functional and covered by tests for store behavior.
- Session files are stored in configured location with clear metadata layout.

Phase 6 — File patch tool  SKILLS.md support
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

Phase 7 — MCP support and sub-agents
Goal: Add an adapter layer to support MCP and plugin-like model routing; implement sub-agent scaffolding.

Steps:
1. Define openai.Adapter interface that exposes ChatOnce / ChatStream / ModelCapabilities. Implement the current OpenAI-compatible client to satisfy it.
2. Implement an MCP adapter that can be selected via configuration; begin with a stub or mock that demonstrates routing behavior.
3. Design a SubAgent interface (Run(context, request) > result) and a manager that can spawn and supervise sub-agents with scoped permissions and timeouts.
4. Provide examples: a code-gen sub-agent that runs in a restricted workspace, a test-runner sub-agent that executes tests with policy guards.

Acceptance criteria:
- Adapter abstraction exists and is used by the agent.
- Sub-agent interface and a minimal implementation exist with tests.

Tests, formatting, linting and PR hygiene

- Run gofmt (gofmt -w) on changed files and run go test ./... locally. Make small commits to keep diffs reviewable.
- Add or update unit tests as part of each phase. Prioritize tests for policy, config precedence, and REPL parsing.
- Consider adding a pre-commit script (optional) to run gofmt and go vet for contributors.

Documentation updates

- Update README.md where user-facing behavior changes (config locations, new subcommands, sessions storage).
- Add brief HOWTOs for:
  - Running with --readonly and --dry-shell
  - Managing sessions
  - Adding a SKILLS.md file
  - Enabling sandboxing (requirements, platform support)

Risk analysis and mitigation

- Large refactors can introduce regressions: mitigate by small incremental moves, frequent tests, and ensuring the project builds between commits.
- Platform differences for sandbox tools: detect availability and fallback to non-sandboxed runner while documenting limitations.
- Dependency bloat (e.g., using viper/cobra/v1): weigh trade-offs  cobra is recommended for CLI UX, viper if full config feature set needed. An explicit config loader may suffice to reduce transitive deps.

Suggested incremental commit sequence (example)

1. Create internal/config with a simple config loader and tests; wire basic struct into cmd/jorin.
2. Move REPL into internal/ui and make startREPL accept io.Reader/io.Writer. Add tests for slash command parsing.
3. Introduce internal/tools and internal/shell with interfaces; replace ad-hoc shell usage with shell.Runner.
4. Move OpenAI client to internal/openai and add Adapter interface.
5. Add session store interface and simple file-backed implementation.
6. Implement CLI subcommands using cobra (or structured pflag) and wire everything together.

Closing notes

- Keep commits small and focused: move code first (preserve behavior), then improve behavior in subsequent commits.
- Prioritize testability and clear interfaces  this pays off by making future features (MCP, sub-agents, SKILLS.md, patching) much easier to implement.
- If you want, I can generate a minimal skeletal set of interface files and TODO stubs to kick-start the refactor (no code will be changed unless you ask).

End of ROADMAP
