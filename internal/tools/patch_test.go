package tools

import (
	"os"
	"testing"
)

func TestParsePatch(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		patch := `--- /dev/null
+++ b/foo.txt
@@ -0,0 +1,1 @@
+hello
`
		p, err := parsePatch(patch)
		if err != nil {
			t.Fatalf("parsePatch() error = %v", err)
		}
		if p.op != opCreate {
			t.Errorf("p.op = %v, want %v", p.op, opCreate)
		}
		if p.filePath != "foo.txt" {
			t.Errorf("p.filePath = %q, want %q", p.filePath, "foo.txt")
		}
		expectedDiff := "+hello"
		if p.diff != expectedDiff {
			t.Errorf("p.diff = %q, want %q", p.diff, expectedDiff)
		}
	})

	t.Run("update", func(t *testing.T) {
		patch := `--- a/foo.txt
+++ b/foo.txt
@@ -1,1 +1,1 @@
-hello
+world
`
		p, err := parsePatch(patch)
		if err != nil {
			t.Fatalf("parsePatch() error = %v", err)
		}
		if p.op != opUpdate {
			t.Errorf("p.op = %v, want %v", p.op, opUpdate)
		}
		if p.filePath != "foo.txt" {
			t.Errorf("p.filePath = %q, want %q", p.filePath, "foo.txt")
		}
		expectedDiff := "-hello\n+world"
		if p.diff != expectedDiff {
			t.Errorf("p.diff = %q, want %q", p.diff, expectedDiff)
		}
	})

	t.Run("delete", func(t *testing.T) {
		patch := `--- a/foo.txt
+++ /dev/null
@@ -1,1 +0,0 @@
-hello
`
		p, err := parsePatch(patch)
		if err != nil {
			t.Fatalf("parsePatch() error = %v", err)
		}
		if p.op != opDelete {
			t.Errorf("p.op = %v, want %v", p.op, opDelete)
		}
		if p.filePath != "foo.txt" {
			t.Errorf("p.filePath = %q, want %q", p.filePath, "foo.txt")
		}
		expectedDiff := "-hello"
		if p.diff != expectedDiff {
			t.Errorf("p.diff = %q, want %q", p.diff, expectedDiff)
		}
	})

	t.Run("invalid patch", func(t *testing.T) {
		_, err := parsePatch("foo")
		if err == nil {
			t.Error("expected an error for invalid patch")
		}
	})

	t.Run("mismatched file paths", func(t *testing.T) {
		patch := `--- a/foo.txt
+++ b/bar.txt
`
		_, err := parsePatch(patch)
		if err == nil {
			t.Error("expected an error for mismatched file paths")
		}
	})
}

func TestApplyPatch_Create(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to change back to wd: %v", err)
		}
	}()

	file := "foo.txt"

	patch := `--- /dev/null
+++ b/foo.txt
@@ -0,0 +1,2 @@
+hello
+world
`

	err = ApplyPatch(patch)
	if err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	expectedContent := "hello\nworld\n"
	if string(content) != expectedContent {
		t.Errorf("content = %q, want %q", string(content), expectedContent)
	}
}

func TestApplyPatch_Update(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to change back to wd: %v", err)
		}
	}()

	file := "foo.txt"
	if err := os.WriteFile(file, []byte("hello\nworld\n"), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	patch := `--- a/foo.txt
+++ b/foo.txt
@@ -1,2 +1,2 @@
-hello
+hi
 world
`

	err = ApplyPatch(patch)
	if err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	expectedContent := "hi\nworld\n"
	if string(content) != expectedContent {
		t.Errorf("content = %q, want %q", string(content), expectedContent)
	}
}

func TestApplyPatch_Delete(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to change back to wd: %v", err)
		}
	}()

	file := "foo.txt"

	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	patch := `--- a/foo.txt
+++ /dev/null
@@ -1,1 +0,0 @@
-hello
`
	err = ApplyPatch(patch)
	if err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("file %q should have been deleted", file)
	}
}
