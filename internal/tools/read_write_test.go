package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dave1010/jorin/internal/types"
)

func TestReadFileMissingPath(t *testing.T) {
	r := Registry()

	if _, err := r["read_file"](map[string]any{}, &types.Policy{}); err == nil {
		t.Fatalf("expected error for missing path")
	}
}

func TestReadFileSuccess(t *testing.T) {
	r := Registry()

	tmp := t.TempDir()
	f := filepath.Join(tmp, "t.txt")
	if err := os.WriteFile(f, []byte("hello tools"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	out, err := r["read_file"](map[string]any{"path": f}, &types.Policy{})
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	if out["text"] != "hello tools" {
		t.Fatalf("unexpected text: %#v", out["text"])
	}
}

func TestWriteFileReadonly(t *testing.T) {
	r := Registry()

	tmp := t.TempDir()
	out, err := r["write_file"](map[string]any{"path": filepath.Join(tmp, "x.txt"), "text": "ok"}, &types.Policy{Readonly: true})
	if err != nil {
		t.Fatalf("write_file returned err: %v", err)
	}
	if out["error"] != "readonly session" {
		t.Fatalf("expected readonly error, got: %#v", out)
	}
}

func TestWriteFileSuccess(t *testing.T) {
	r := Registry()

	tmp := t.TempDir()
	out, err := r["write_file"](map[string]any{"path": filepath.Join(tmp, "x.txt"), "text": "ok"}, &types.Policy{})
	if err != nil {
		t.Fatalf("write_file failed: %v", err)
	}
	if ok, _ := out["ok"].(bool); !ok {
		t.Fatalf("expected ok true, got %#v", out)
	}
}
