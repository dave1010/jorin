package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const systemPrompt = `You are a coding agent, designed to call tools to complete tasks.
Respond either with a normal assistant message, or with tool calls (function calling).
Prefer small, auditable steps. Read before you write. Summarize long outputs.`

// loadSystemPrompt returns the base system prompt and, if an AGENTS.md file
// exists in the current working directory, appends its contents preceded by
// "Project-specific instructions:".
func loadSystemPrompt() string {
	sp := systemPrompt
	if _, err := os.Stat("AGENTS.md"); err == nil {
		if b, err := os.ReadFile("AGENTS.md"); err == nil {
			sp = sp + "\n\nProject-specific instructions:\n" + string(b)
		}
	}
	return sp
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
	fmt.Println("agent> (Ctrl-D to exit)")
	msgs := []Message{{Role: "system", Content: loadSystemPrompt()}}
	for {
		fmt.Print("> ")
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
			fmt.Println("ERR:", err)
			continue
		}
		fmt.Println(out)
	}
}
