package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/app"
	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/version"
)

type Config struct {
	model           string
	repl            bool
	readonly        bool
	dryShell        bool
	allow           []string
	deny            []string
	cwd             string
	promptFlag      bool
	promptFileFlag  bool
	ralph           bool
	ralphMaxTries   int
	versionFlag     bool
	useResponsesAPI bool
}

func parseFlags() Config {
	model := flag.String("model", "gpt-5-mini", "Model ID")
	repl := flag.Bool("repl", false, "Interactive REPL")
	readonly := flag.Bool("readonly", false, "Disallow write_file")
	dry := flag.Bool("dry-shell", false, "Do not execute shell commands")
	allow := multi("allow", "Allowlist substring for shell (repeatable)")
	deny := multi("deny", "Denylist substring for shell (repeatable)")
	cwd := flag.String("cwd", "", "Working directory for tools")
	promptFlag := flag.Bool("prompt", false, "Treat first argument as prompt text")
	promptFileFlag := flag.Bool("prompt-file", false, "Treat first argument as a prompt file")
	ralph := flag.Bool("ralph", false, "Enable Ralph Wiggum loop instructions")
	ralphMaxTries := flag.Int("ralph-max-tries", 8, "Maximum Ralph Wiggum loop iterations")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	useResponsesAPI := flag.Bool("use-responses-api", false, "Use the new OpenAI Responses API instead of Chat Completions")
	flag.Parse()

	return Config{
		model:           *model,
		repl:            *repl,
		readonly:        *readonly,
		dryShell:        *dry,
		allow:           *allow,
		deny:            *deny,
		cwd:             *cwd,
		promptFlag:      *promptFlag,
		promptFileFlag:  *promptFileFlag,
		ralph:           *ralph,
		ralphMaxTries:   *ralphMaxTries,
		versionFlag:     *versionFlag,
		useResponsesAPI: *useResponsesAPI,
	}
}

func handlePreflight(cli Config) {
	if cli.versionFlag {
		fmt.Println(version.Version)
		os.Exit(0)
	}
	if cli.promptFlag && cli.promptFileFlag {
		fmt.Fprintln(os.Stderr, "ERR: flag --prompt and --prompt-file cannot be used together")
		os.Exit(2)
	}
	if cli.ralph {
		prompt.EnableRalph()
	}
	if cli.ralphMaxTries < 1 {
		fmt.Fprintln(os.Stderr, "ERR: flag --ralph-max-tries must be at least 1")
		os.Exit(2)
	}
}

func resolvePromptMode(promptFlag bool, promptFileFlag bool) promptMode {
	if promptFlag {
		return promptModeText
	}
	if promptFileFlag {
		return promptModeFile
	}
	return promptModeAuto
}

func exitWithError(err error) {
	if err == app.ErrMissingPrompt {
		fmt.Fprintln(os.Stderr, "Provide a prompt or use --repl")
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "ERR:", err)
	os.Exit(1)
}

// multi flag helpers
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }
func (m *multiFlag) Type() string       { return "stringSlice" }
func multi(name, usage string) *multiFlag {
	var v multiFlag
	flag.Var(&v, name, usage)
	return &v
}
