package tools

import (
	"os"
	"strings"
	testing "testing"
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
		if len(p.hunks) != 1 {
			t.Fatalf("len(p.hunks) = %d, want 1", len(p.hunks))
		}
		expectedLines := []string{`+hello`}
		for i, line := range expectedLines {
			if p.hunks[0].lines[i] != line {
				t.Errorf("p.hunks[0].lines[%d] = %q, want %q", i, p.hunks[0].lines[i], line)
			}
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
		if len(p.hunks) != 1 {
			t.Fatalf("len(p.hunks) = %d, want 1", len(p.hunks))
		}
		expectedLines := []string{`-hello`, `+world`}
		for i, line := range expectedLines {
			if p.hunks[0].lines[i] != line {
				t.Errorf("p.hunks[0].lines[%d] = %q, want %q", i, p.hunks[0].lines[i], line)
			}
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
		if len(p.hunks) != 1 {
			t.Fatalf("len(p.hunks) = %d, want 1", len(p.hunks))
		}
		expectedLines := []string{`-hello`}
		for i, line := range expectedLines {
			if p.hunks[0].lines[i] != line {
				t.Errorf("p.hunks[0].lines[%d] = %q, want %q", i, p.hunks[0].lines[i], line)
			}
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

	t.Run("patch with leading conversational text", func(t *testing.T) {
		patch := `Here is the patch you requested:

--- a/foo.txt
+++ b/foo.txt
@@ -1,1 +1,1 @@
-hello
+world
`
		p, err := parsePatch(patch)
		if err != nil {
			t.Fatalf("parsePatch() error = %v", err)
		}
		if p.filePath != "foo.txt" {
			t.Errorf("p.filePath = %q, want %q", p.filePath, "foo.txt")
		}
		if len(p.hunks) != 1 {
			t.Fatalf("len(p.hunks) = %d, want 1", len(p.hunks))
		}
		expectedLines := []string{`-hello`, `+world`}
		for i, line := range expectedLines {
			if p.hunks[0].lines[i] != line {
				t.Errorf("p.hunks[0].lines[%d] = %q, want %q", i, p.hunks[0].lines[i], line)
			}
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

func TestApplyPatch_MultipleHunks(t *testing.T) {
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

	file := "bar.txt"
	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	patch := `--- a/bar.txt
+++ b/bar.txt
@@ -2,1 +2,1 @@
-line 2
+line TWO
@@ -4,1 +4,2 @@
-line 4
+line FOUR
+line FOUR.5
`

	err = ApplyPatch(patch)
	if err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}

	actual, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	expected := "line 1\nline TWO\nline 3\nline FOUR\nline FOUR.5\nline 5\n"
	if string(actual) != expected {
		t.Errorf("content = %q, want %q", string(actual), expected)
	}
}

func TestApplyPatch_ReproIssues(t *testing.T) {
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

	file := "repro.txt"
	initialContent := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if err := os.WriteFile(file, []byte(initialContent), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	// Issue 1: Invalid hunk header "@@"
	t.Run("InvalidHunkHeader", func(t *testing.T) {
		patch := `--- a/repro.txt
+++ b/repro.txt
@@
-line 1
+LINE 1
`
		err := ApplyPatch(patch)
		if err == nil {
			t.Error("Expected error for invalid hunk header '@@', got nil")
		} else {
			expected := "malformed hunk header"
			if !strings.Contains(err.Error(), expected) {
				t.Errorf("Error message expected to contain %q, got %q", expected, err.Error())
			}
		}
	})

	// Issue 2: Mismatched context (fuzzy match needed)
	// The patch claims the hunk starts at line 1, but "line 3" is actually at line 3.
	t.Run("MismatchedContext_FuzzyMatchSuccess", func(t *testing.T) {
		// Reset file content
		if err := os.WriteFile(file, []byte(initialContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		patch := `--- a/repro.txt
+++ b/repro.txt
@@ -1,1 +1,1 @@
-line 3
+LINE 3
`
		err := ApplyPatch(patch)
		if err != nil {
			t.Errorf("Expected fuzzy match to succeed, got error: %v", err)
		}

		content, _ := os.ReadFile(file)
		expected := "line 1\nline 2\nLINE 3\nline 4\nline 5\n"
		if string(content) != expected {
			t.Errorf("File content mismatch. Got:\n%s\nWant:\n%s", string(content), expected)
		}
	})

	// Issue 3: Verify the error message when match truly fails
	t.Run("FuzzyMatchFail", func(t *testing.T) {
		// Reset file content
		if err := os.WriteFile(file, []byte(initialContent), 0644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		patch := `--- a/repro.txt
+++ b/repro.txt
@@ -1,1 +1,1 @@
-line NOT FOUND
+LINE 3
`
		err := ApplyPatch(patch)
		if err == nil {
			t.Error("Expected error for hunk that cannot be found")
		} else {
			if !strings.Contains(err.Error(), "hunk context not found") {
				t.Errorf("Error message should explain failure. Got: %v", err)
			}
			if !strings.Contains(err.Error(), "Expected first") {
				t.Errorf("Error message should show expectations. Got: %v", err)
			}
		}
	})
}
