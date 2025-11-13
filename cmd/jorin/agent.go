package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const systemPrompt = `You are a coding agent, designed to call tools to complete tasks.
Respond either with a normal assistant message, or with tool calls (function calling).
Prefer small, auditable steps. Read before you write. Don't suggest extra work.`

// loadSystemPrompt returns the base system prompt and, if an AGENTS.md file
// exists in the current working directory, appends its contents preceded by
// "Project-specific instructions:". It also appends runtime environment
// context (uname/runtime info, working directory, presence of a git
// repository, and a short list of handy tools found on PATH).
func loadSystemPrompt() string {
	sp := systemPrompt
	if _, err := os.Stat("AGENTS.md"); err == nil {
		if b, err := os.ReadFile("AGENTS.md"); err == nil {
			sp = sp + "\n\nProject-specific instructions:\n" + string(b)
		}
	}
	// Append runtime context
	ctx := runtimeContext()
	if ctx != "" {
		sp = sp + "\n\nRuntime environment:\n" + ctx
	}
	return sp
}

func runtimeContext() string {
	parts := []string{}
	// uname -a if available
	if out, err := exec.Command("uname", "-a").Output(); err == nil {
		parts = append(parts, strings.TrimSpace(string(out)))
	} else {
		// fallback to GOOS/GOARCH
		parts = append(parts, "OS: "+runtime.GOOS+" "+runtime.GOARCH)
	}
	// working directory
	if wd, err := os.Getwd(); err == nil {
		parts = append(parts, "PWD: "+wd)
	}
	// git repo presence
	if _, err := os.Stat(".git"); err == nil {
		parts = append(parts, "Git repository: yes (.git exists)")
	} else {
		parts = append(parts, "Git repository: no (.git not found)")
	}
	// check for a few handy tools
	tools := []string{"ag", "rg", "git", "gh", "go", "gofmt", "docker", "fzf", "python", "python3", "php", "curl", "wget"}
	found := []string{}
	for _, t := range tools {
		if _, err := exec.LookPath(t); err == nil {
			found = append(found, t+" ")
		}
	}
	if len(found) > 0 {
		parts = append(parts, "Tools on PATH (others will exist too): "+strings.Join(found, ", "))
	} else {
		parts = append(parts, "Tools on PATH: none of "+strings.Join(tools, ", "))
	}
	return strings.Join(parts, "\n")
}

func runAgent(model string, userPrompt string, pol *Policy) (string, error) {
	msgs := []Message{
		{Role: "system", Content: loadSystemPrompt()},
		{Role: "user", Content: userPrompt},
	}
	_, out, err := chatSession(model, msgs, pol)
	return out, err
}

// kept for REPL support in main
func startREPL(model string, pol *Policy) {
	in := bufio.NewScanner(os.Stdin)
	fmt.Println(headerStyleStr("jorin> (Ctrl-D to exit)"))
	msgs := []Message{{Role: "system", Content: loadSystemPrompt()}}
	for {
		fmt.Print(promptStyleStr("> "))
		if !in.Scan() {
			break
		}
		q := strings.TrimSpace(in.Text())
		if q == "" {
			continue
		}
		msgs = append(msgs, Message{Role: "user", Content: q})
		var out string
		var err error
		msgs, out, err = chatSession(model, msgs, pol)
		if err != nil {
			fmt.Println(errorStyleStr("ERR:"), err)
			continue
		}
		fmt.Println(infoStyleStr(out))
	}
}
