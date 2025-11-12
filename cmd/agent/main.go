package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name string `json:"name"`
		Args string `json:"arguments"`
	} `json:"function"`
}

type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"` // JSON Schema
}

type ChatRequest struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	Tools       []Tool      `json:"tools,omitempty"`
	ToolChoice  interface{} `json:"tool_choice,omitempty"` // "auto"
	Temperature float32     `json:"temperature,omitempty"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// --- tool registry ---

type ToolExec func(args map[string]any, cfg *Policy) (map[string]any, error)

type Policy struct {
	Readonly bool
	DryShell bool
	Allow    []string
	Deny     []string
	CWD      string
}

func schema(s string) json.RawMessage { return json.RawMessage([]byte(s)) }

func toolsManifest() (list []Tool) {
	return []Tool{
		{Type: "function", Function: ToolFunction{
			Name:        "shell",
			Description: "Execute a shell command; returns stdout/stderr/returncode. Use cautiously.",
			Parameters:  schema(`{"type":"object","properties":{"cmd":{"type":"string"}},"required":["cmd"]}`),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "read_file",
			Description: "Read a UTF-8 text file and return contents (truncated).",
			Parameters:  schema(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "write_file",
			Description: "Write UTF-8 text to a file (creates/overwrites). Disabled in readonly.",
			Parameters:  schema(`{"type":"object","properties":{"path":{"type":"string"},"text":{"type":"string"}},"required":["path","text"]}`),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "http_get",
			Description: "Fetch URL and return body (text).",
			Parameters:  schema(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`),
		}},
	}
}

func registry() map[string]ToolExec {
	return map[string]ToolExec{
		"shell": func(args map[string]any, p *Policy) (map[string]any, error) {
			cmdStr, _ := args["cmd"].(string)
			if cmdStr == "" {
				return nil, errors.New("missing cmd")
			}
			for _, d := range p.Deny {
				if strings.Contains(cmdStr, d) {
					return map[string]any{"error": "denied by policy"}, nil
				}
			}
			if len(p.Allow) > 0 {
				ok := false
				for _, a := range p.Allow {
					if strings.Contains(cmdStr, a) {
						ok = true
						break
					}
				}
				if !ok {
					return map[string]any{"error": "not allowed by policy"}, nil
				}
			}
			if p.DryShell {
				return map[string]any{"dry_run": true, "cmd": cmdStr}, nil
			}
			c := exec.Command("bash", "-lc", cmdStr)
			c.Dir = p.CWD
			var out bytes.Buffer
			var errb bytes.Buffer
			c.Stdout = &out
			c.Stderr = &errb
			cErr := c.Run()
			rc := 0
			if cErr != nil {
				if ee, ok := cErr.(*exec.ExitError); ok {
					rc = ee.ExitCode()
				} else {
					rc = 1
				}
			}
			return map[string]any{
				"returncode": rc,
				"stdout":     tail(out.String(), 8000),
				"stderr":     tail(errb.String(), 8000),
			}, nil
		},
		"read_file": func(args map[string]any, p *Policy) (map[string]any, error) {
			path, _ := args["path"].(string)
			if path == "" {
				return nil, errors.New("missing path")
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return map[string]any{"error": err.Error()}, nil
			}
			txt := string(b)
			trunc := false
			if len(txt) > 200_000 {
				txt = txt[:200_000]
				trunc = true
			}
			return map[string]any{"text": txt, "truncated": trunc}, nil
		},
		"write_file": func(args map[string]any, p *Policy) (map[string]any, error) {
			if p.Readonly {
				return map[string]any{"error": "readonly session"}, nil
			}
			path, _ := args["path"].(string)
			text, _ := args["text"].(string)
			if path == "" {
				return nil, errors.New("missing path")
			}
			if err := os.MkdirAll(dirOrDot(path), 0o755); err != nil {
				return map[string]any{"error": err.Error()}, nil
			}
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				return map[string]any{"error": err.Error()}, nil
			}
			return map[string]any{"ok": true, "bytes": len(text)}, nil
		},
		"http_get": func(args map[string]any, p *Policy) (map[string]any, error) {
			url, _ := args["url"].(string)
			if url == "" {
				return nil, errors.New("missing url")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return map[string]any{"error": err.Error()}, nil
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 8000))
			return map[string]any{"status": resp.StatusCode, "body": string(b)}, nil
		},
	}
}

func dirOrDot(p string) string {
	d := filepath.Dir(p)
	if d == "" || d == "." {
		return "."
	}
	return d
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// helper: trim and single-line preview
func preview(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// --- provider: OpenAI-compatible ---

func openAIBase() string {
	if b := os.Getenv("OPENAI_BASE_URL"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://api.openai.com"
}

func chatOnce(model string, msgs []Message, tools []Tool) (*ChatResponse, error) {
	body := ChatRequest{
		Model:      model,
		Messages:   msgs,
		Tools:      tools,
		ToolChoice: "auto",
	}
	j, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", openAIBase()+"/v1/chat/completions", bytes.NewReader(j))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(b))
	}
	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// chatSession runs the model starting from msgs and executes any tool calls until the assistant
// returns a non-tool-calling message. It returns the updated messages (with assistant/tool
// messages appended) and the assistant's final content.
func chatSession(model string, msgs []Message, pol *Policy) ([]Message, string, error) {
	tools := toolsManifest()
	reg := registry()
	for i := 0; i < 16; i++ {
		resp, err := chatOnce(model, msgs, tools)
		if err != nil {
			return msgs, "", err
		}
		if len(resp.Choices) == 0 {
			return msgs, "", errors.New("no choices")
		}
		ch := resp.Choices[0]
		cm := ch.Message

		// append assistant message
		msgs = append(msgs, cm)

		// If tool calls present, execute them serially and loop
		if len(cm.ToolCalls) > 0 {
			for _, tc := range cm.ToolCalls {
				// show which tool is being called (trimmed)
				fmt.Fprintln(os.Stderr, "CALL TOOL:", tc.Function.Name, preview(tc.Function.Args, 200))
				fn := reg[tc.Function.Name]
				if fn == nil {
					msgs = append(msgs, Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"unknown tool"}`,
					})
					continue
				}
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Args), &args); err != nil {
					msgs = append(msgs, Message{
						Role:       "tool",
						Name:       tc.Function.Name,
						ToolCallID: tc.ID,
						Content:    `{"error":"bad arguments"}`,
					})
					continue
				}
				out, _ := fn(args, pol)
				b, _ := json.Marshal(out)
				msgs = append(msgs, Message{
					Role:       "tool",
					Name:       tc.Function.Name,
					ToolCallID: tc.ID,
					Content:    string(b),
				})
			}
			// continue the loop so model sees tool outputs and can respond
			continue
		}

		// otherwise final assistant content (no tool calls)
		return msgs, cm.Content, nil
	}
	return msgs, "", errors.New("max turns reached")
}

// --- agent loop ---

const systemPrompt = `You are a coding agent, designed to call tools to complete tasks.
Respond either with a normal assistant message, or with tool calls (function calling).
Prefer small, auditable steps. Read before you write. Summarize long outputs.`

func runAgent(model string, userPrompt string, pol *Policy) (string, error) {
	msgs := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	_, out, err := chatSession(model, msgs, pol)
	return out, err
}

// --- CLI ---

func main() {
	model := flag.String("model", "gpt-5-mini", "Model ID")
	repl := flag.Bool("repl", false, "Interactive REPL")
	readonly := flag.Bool("readonly", false, "Disallow write_file")
	dry := flag.Bool("dry-shell", false, "Do not execute shell commands")
	allow := multi("allow", "Allowlist substring for shell (repeatable)")
	deny := multi("deny", "Denylist substring for shell (repeatable)")
	cwd := flag.String("cwd", "", "Working directory for tools")
	flag.Parse()

	pol := &Policy{Readonly: *readonly, DryShell: *dry, Allow: *allow, Deny: *deny, CWD: *cwd}

	if *repl {
		in := bufio.NewScanner(os.Stdin)
		fmt.Println("agent> (Ctrl-D to exit)")
		// start conversation with system prompt
		msgs := []Message{{Role: "system", Content: systemPrompt}}
		for {
			fmt.Print("> ")
			if !in.Scan() {
				break
			}
			q := strings.TrimSpace(in.Text())
			if q == "" {
				continue
			}
			// append user message and continue the same conversation
			msgs = append(msgs, Message{Role: "user", Content: q})
			var out string
			var err error
			msgs, out, err = chatSession(*model, msgs, pol)
			if err != nil {
				fmt.Println("ERR:", err)
				continue
			}
			fmt.Println(out)
		}
		return
	}

	prompt := strings.Join(flag.Args(), " ")
	if strings.TrimSpace(prompt) == "" {
		fmt.Fprintln(os.Stderr, "Provide a prompt or use --repl")
		os.Exit(2)
	}
	out, err := runAgent(*model, prompt, pol)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

type multiFlag []string

func (m *multiFlag) String() string       { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error   { *m = append(*m, v); return nil }
func multi(name, usage string) *multiFlag { var v multiFlag; flag.Var(&v, name, usage); return &v }
