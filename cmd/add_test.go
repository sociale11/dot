package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTest creates a fake root and dot dir, returns their paths.
func setupTest(t *testing.T) (root, dot string) {
	t.Helper()
	root = t.TempDir()
	dot = filepath.Join(root, ".local/share/dot")
	if err := os.MkdirAll(dot, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return root, dot
}

// writeFile is a test helper for creating files with content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestAdd_SimpleFile(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "export PATH=/foo")

	if err := add(original, root, dot); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Original location should now be a symlink.
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink, got mode %v", original, info.Mode())
	}

	// Symlink should point into dot.
	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	expected := filepath.Join(dot, ".zshrc")
	if target != expected {
		t.Errorf("symlink target: got %s, want %s", target, expected)
	}

	// Content should be preserved and readable through the symlink.
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "export PATH=/foo" {
		t.Errorf("content: got %q, want %q", got, "export PATH=/foo")
	}
}

func TestAdd_NestedFile(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".config/nvim/init.lua")
	writeFile(t, original, "vim.opt.number = true")

	if err := add(original, root, dot); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	tracked := filepath.Join(dot, ".config/nvim/init.lua")
	if _, err := os.Stat(tracked); err != nil {
		t.Errorf("tracked file not created: %v", err)
	}
}

func TestAdd_RejectsPathOutsideRoot(t *testing.T) {
	root, dot := setupTest(t)

	outside := filepath.Join(t.TempDir(), "evil")
	writeFile(t, outside, "nope")

	err := add(outside, root, dot)
	if err == nil {
		t.Fatal("expected error for path outside root, got nil")
	}
	if !strings.Contains(err.Error(), "not under root") {
		t.Errorf("error message: got %q, want it to mention 'not under root'", err)
	}
}

func TestAdd_RefusesToReAddTrackedFile(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "important config")

	if err := add(original, root, dot); err != nil {
		t.Fatalf("first add: %v", err)
	}

	err := add(original, root, dot)
	if err == nil {
		t.Fatal("expected re-add to fail, got nil")
	}

	// Content must survive the failed re-add.
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read after failed re-add: %v", err)
	}
	if string(got) != "important config" {
		t.Errorf("content was modified during failed re-add: got %q", got)
	}
}

func TestAdd_RefusesForeignSymlink(t *testing.T) {
	root, dot := setupTest(t)

	elsewhere := filepath.Join(t.TempDir(), "elsewhere")
	writeFile(t, elsewhere, "someone else's file")

	original := filepath.Join(root, ".zshrc")
	if err := os.Symlink(elsewhere, original); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	err := add(original, root, dot)
	if err == nil {
		t.Fatal("expected add to refuse foreign symlink, got nil")
	}

	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != elsewhere {
		t.Errorf("symlink was modified: got %s, want %s", target, elsewhere)
	}
}

func TestRestore_RoundTrip(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "original content")

	if err := add(original, root, dot); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := restore(original, root, dot); err != nil {
		t.Fatalf("restore: %v", err)
	}

	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected %s to be a regular file, got symlink", original)
	}

	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "original content" {
		t.Errorf("content: got %q, want %q", got, "original content")
	}

	tracked := filepath.Join(dot, ".zshrc")
	if _, err := os.Stat(tracked); !os.IsNotExist(err) {
		t.Errorf("tracked file should be removed, got err: %v", err)
	}
}

func TestRestore_RejectsNonSymlink(t *testing.T) {
	root, dot := setupTest(t)

	regular := filepath.Join(root, "regular.txt")
	writeFile(t, regular, "just a file")

	err := restore(regular, root, dot)
	if err == nil {
		t.Fatal("expected error for non-symlink, got nil")
	}
	if !strings.Contains(err.Error(), "not a symlink") {
		t.Errorf("error message: got %q, want it to mention 'not a symlink'", err)
	}
}

func TestRestore_RejectsSymlinkToElsewhere(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	elsewhere := filepath.Join(t.TempDir(), "other")
	writeFile(t, elsewhere, "not ours")
	if err := os.Symlink(elsewhere, original); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	err := restore(original, root, dot)
	if err == nil {
		t.Fatal("expected error for foreign symlink, got nil")
	}
}

// --- Directory (whole-dir symlink) tests ---

func TestAddPath_DirectoryMovesAndSymlinks(t *testing.T) {
	root, dot := setupTest(t)

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "a")
	writeFile(t, filepath.Join(dir, "lua/plugins/foo.lua"), "b")

	if err := addPath(dir, root, dot); err != nil {
		t.Fatalf("addPath directory: %v", err)
	}

	// Original path should be a symlink to dot.
	info, err := os.Lstat(dir)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink, got mode %v", dir, info.Mode())
	}

	target, err := os.Readlink(dir)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	expected := filepath.Join(dot, ".config/nvim")
	if target != expected {
		t.Errorf("symlink target: got %s, want %s", target, expected)
	}

	// Files should be readable through the symlink.
	got, err := os.ReadFile(filepath.Join(dir, "init.lua"))
	if err != nil {
		t.Fatalf("read through symlink: %v", err)
	}
	if string(got) != "a" {
		t.Errorf("content: got %q, want %q", got, "a")
	}

	gotNested, err := os.ReadFile(filepath.Join(dir, "lua/plugins/foo.lua"))
	if err != nil {
		t.Fatalf("read nested through symlink: %v", err)
	}
	if string(gotNested) != "b" {
		t.Errorf("nested content: got %q, want %q", gotNested, "b")
	}
}

func TestAddPath_DirectoryRejectsAlreadyTracked(t *testing.T) {
	root, dot := setupTest(t)

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "a")

	if err := addPath(dir, root, dot); err != nil {
		t.Fatalf("first add: %v", err)
	}

	// Second add should fail — it's already a symlink into dot.
	err := addPath(dir, root, dot)
	if err == nil {
		t.Fatal("expected re-add to fail, got nil")
	}

	// Content must still be accessible.
	got, err := os.ReadFile(filepath.Join(dir, "init.lua"))
	if err != nil {
		t.Fatalf("read after failed re-add: %v", err)
	}
	if string(got) != "a" {
		t.Errorf("content: got %q, want %q", got, "a")
	}
}

func TestAddPath_DirectoryOutsideRootFails(t *testing.T) {
	root, dot := setupTest(t)

	outside := filepath.Join(t.TempDir(), "somedir")
	writeFile(t, filepath.Join(outside, "file.txt"), "nope")

	err := addPath(outside, root, dot)
	if err == nil {
		t.Fatal("expected error for directory outside root, got nil")
	}
}
