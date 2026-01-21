package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestShellDryRun(t *testing.T) {
	r := Registry()

	out, err := r["shell"](map[string]any{"cmd": "echo hi"}, &types.Policy{DryShell: true})
	if err != nil {
		t.Fatalf("shell dry failed: %v", err)
	}
	if dr, _ := out["dry_run"].(bool); !dr {
		t.Fatalf("expected dry_run true")
	}
	if out["cmd"].(string) != "echo hi" {
		t.Fatalf("unexpected cmd in dry run: %#v", out)
	}
}

func TestShellCommand(t *testing.T) {
	r := Registry()

	out, err := r["shell"](map[string]any{"cmd": "echo -n OK"}, &types.Policy{})
	if err != nil {
		t.Fatalf("shell failed: %v", err)
	}
	if s, _ := out["stdout"].(string); s != "OK" {
		t.Fatalf("unexpected stdout: %q", s)
	}
}

func TestShellCWD(t *testing.T) {
	r := Registry()

	tmp := t.TempDir()
	script := filepath.Join(tmp, "pw.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho -n $(pwd) > outpwd"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if _, err := r["shell"](map[string]any{"cmd": "./pw.sh"}, &types.Policy{CWD: tmp}); err != nil {
		t.Fatalf("shell CWD failed: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(tmp, "outpwd"))
	if err != nil {
		t.Fatalf("read outpwd: %v", err)
	}
	if string(b) != tmp {
		t.Fatalf("expected %q got %q", tmp, string(b))
	}
}
