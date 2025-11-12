package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

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
			Description: "Execute a shell command; returns stdout/stderr/returncode. Use cautiously if commands may be destructive.",
			Parameters:  schema(`{"type":"object","properties":{"cmd":{"type":"string"}},"required":["cmd"]}`),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "read_file",
			Description: "Read a UTF-8 text file and return contents. Very long files will be truncated.",
			Parameters:  schema(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "write_file",
			Description: "Write UTF-8 text to a file (creates/overwrites).",
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

func preview(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
