package tools

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

	"github.com/dave1010/jorin/internal/types"
)

const (
	maxReadFileBytes   = 200_000
	maxToolOutputBytes = 8000
	httpTimeout        = 15 * time.Second
)

// ToolExec is a function that executes a tool given args and a policy.
type ToolExec func(args map[string]any, cfg *types.Policy) (map[string]any, error)

func schema(s string) json.RawMessage { return json.RawMessage([]byte(s)) }

func ToolsManifest() (list []types.Tool) {
	return []types.Tool{
		{Type: "function", Function: types.ToolFunction{
			Name:        "shell",
			Description: "Execute a shell command; returns stdout/stderr/returncode. Use cautiously if commands may be destructive.",
			Parameters:  schema(`{"type":"object","properties":{"cmd":{"type":"string"}},"required":["cmd"]}`),
		}},
		{Type: "function", Function: types.ToolFunction{
			Name:        "read_file",
			Description: "Read a UTF-8 text file and return contents. Very long files will be truncated.",
			Parameters:  schema(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`),
		}},
		{Type: "function", Function: types.ToolFunction{
			Name:        "write_file",
			Description: "Write UTF-8 text to a file (creates/overwrites).",
			Parameters:  schema(`{"type":"object","properties":{"path":{"type":"string"},"text":{"type":"string"}},"required":["path","text"]}`),
		}},
		{Type: "function", Function: types.ToolFunction{
			Name:        "http_get",
			Description: "Fetch URL and return body (text).",
			Parameters:  schema(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`),
		}},
	}
}

func Registry() map[string]ToolExec {
	return map[string]ToolExec{
		"shell":      shellToolExec,
		"read_file":  readFileToolExec,
		"write_file": writeFileToolExec,
		"http_get":   httpGetToolExec,
	}
}

func shellToolExec(args map[string]any, p *types.Policy) (map[string]any, error) {
	cmdStr, _ := args["cmd"].(string)
	if cmdStr == "" {
		return nil, errors.New("missing cmd")
	}
	if allowed, reason := checkShellPolicy(cmdStr, p); !allowed {
		return map[string]any{"error": reason}, nil
	}
	if p.DryShell {
		return map[string]any{"dry_run": true, "cmd": cmdStr}, nil
	}
	cmd := exec.Command("bash", "-lc", cmdStr)
	cmd.Dir = p.CWD
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	runErr := cmd.Run()
	return map[string]any{
		"returncode": exitCode(runErr),
		"stdout":     Tail(out.String(), maxToolOutputBytes),
		"stderr":     Tail(errb.String(), maxToolOutputBytes),
	}, nil
}

func checkShellPolicy(cmdStr string, p *types.Policy) (bool, string) {
	for _, d := range p.Deny {
		if strings.Contains(cmdStr, d) {
			return false, "denied by policy"
		}
	}
	if len(p.Allow) == 0 {
		return true, ""
	}
	for _, a := range p.Allow {
		if strings.Contains(cmdStr, a) {
			return true, ""
		}
	}
	return false, "not allowed by policy"
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return 1
}

func readFileToolExec(args map[string]any, _ *types.Policy) (map[string]any, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return nil, errors.New("missing path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	txt := string(b)
	if len(txt) > maxReadFileBytes {
		return map[string]any{"text": txt[:maxReadFileBytes], "truncated": true}, nil
	}
	return map[string]any{"text": txt, "truncated": false}, nil
}

func writeFileToolExec(args map[string]any, p *types.Policy) (map[string]any, error) {
	if p.Readonly {
		return map[string]any{"error": "readonly session"}, nil
	}
	path, _ := args["path"].(string)
	text, _ := args["text"].(string)
	if path == "" {
		return nil, errors.New("missing path")
	}
	if err := os.MkdirAll(DirOrDot(path), 0o755); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	return map[string]any{"ok": true, "bytes": len(text)}, nil
}

func httpGetToolExec(args map[string]any, _ *types.Policy) (map[string]any, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return nil, errors.New("missing url")
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, maxToolOutputBytes))
	return map[string]any{"status": resp.StatusCode, "body": string(b)}, nil
}

func DirOrDot(p string) string {
	d := filepath.Dir(p)
	if d == "" || d == "." {
		return "."
	}
	return d
}

func Tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func Preview(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
