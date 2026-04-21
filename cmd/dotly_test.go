package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTest creates a fake root and dotly dir, returns their paths.
func setupTest(t *testing.T) (root, dotly string) {
	t.Helper()
	root = t.TempDir()
	dotly = filepath.Join(root, ".local/share/dotly")
	if err := os.MkdirAll(dotly, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return root, dotly
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
	root, dotly := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "export PATH=/foo")

	if err := add(original, root, dotly); err != nil {
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

	// Symlink should point into dotly.
	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	expected := filepath.Join(dotly, ".zshrc")
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
	root, dotly := setupTest(t)

	original := filepath.Join(root, ".config/nvim/init.lua")
	writeFile(t, original, "vim.opt.number = true")

	if err := add(original, root, dotly); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	tracked := filepath.Join(dotly, ".config/nvim/init.lua")
	if _, err := os.Stat(tracked); err != nil {
		t.Errorf("tracked file not created: %v", err)
	}
}

func TestAdd_RejectsPathOutsideRoot(t *testing.T) {
	root, dotly := setupTest(t)

	// A path clearly outside root.
	outside := filepath.Join(t.TempDir(), "evil")
	writeFile(t, outside, "nope")

	err := add(outside, root, dotly)
	if err == nil {
		t.Fatal("expected error for path outside root, got nil")
	}
	if !strings.Contains(err.Error(), "not under root") {
		t.Errorf("error message: got %q, want it to mention 'not under root'", err)
	}
}

func TestAdd_RefusesToReAddTrackedFile(t *testing.T) {
	root, dotly := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "important config")

	// First add: should succeed.
	if err := add(original, root, dotly); err != nil {
		t.Fatalf("first add: %v", err)
	}

	// Second add: must fail, and the content must be preserved.
	err := add(original, root, dotly)
	if err == nil {
		t.Fatal("expected re-add to fail, got nil")
	}

	// Critically: the file content must survive the failed re-add.
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read after failed re-add: %v", err)
	}
	if string(got) != "important config" {
		t.Errorf("content was modified during failed re-add: got %q", got)
	}
}

func TestAdd_RefusesForeignSymlink(t *testing.T) {
	root, dotly := setupTest(t)

	// A symlink at original location pointing somewhere outside DOTLY.
	elsewhere := filepath.Join(t.TempDir(), "elsewhere")
	writeFile(t, elsewhere, "someone else's file")

	original := filepath.Join(root, ".zshrc")
	if err := os.Symlink(elsewhere, original); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	err := add(original, root, dotly)
	if err == nil {
		t.Fatal("expected add to refuse foreign symlink, got nil")
	}

	// The foreign symlink must still exist, pointing where it did.
	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != elsewhere {
		t.Errorf("symlink was modified: got %s, want %s", target, elsewhere)
	}
}

func TestRestore_RoundTrip(t *testing.T) {
	root, dotly := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "original content")

	if err := add(original, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := restore(original, root, dotly); err != nil {
		t.Fatalf("restore: %v", err)
	}

	// Should be a regular file again, not a symlink.
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected %s to be a regular file, got symlink", original)
	}

	// Content preserved.
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "original content" {
		t.Errorf("content: got %q, want %q", got, "original content")
	}

	// Tracked copy should be gone.
	tracked := filepath.Join(dotly, ".zshrc")
	if _, err := os.Stat(tracked); !os.IsNotExist(err) {
		t.Errorf("tracked file should be removed, got err: %v", err)
	}
}

func TestRestore_RejectsNonSymlink(t *testing.T) {
	root, dotly := setupTest(t)

	// A regular file, never added.
	regular := filepath.Join(root, "regular.txt")
	writeFile(t, regular, "just a file")

	err := restore(regular, root, dotly)
	if err == nil {
		t.Fatal("expected error for non-symlink, got nil")
	}
	if !strings.Contains(err.Error(), "not a symlink") {
		t.Errorf("error message: got %q, want it to mention 'not a symlink'", err)
	}
}

func TestRestore_RejectsSymlinkToElsewhere(t *testing.T) {
	root, dotly := setupTest(t)

	// A symlink that points somewhere other than into dotly.
	original := filepath.Join(root, ".zshrc")
	elsewhere := filepath.Join(t.TempDir(), "other")
	writeFile(t, elsewhere, "not ours")
	if err := os.Symlink(elsewhere, original); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	err := restore(original, root, dotly)
	if err == nil {
		t.Fatal("expected error for foreign symlink, got nil")
	}
}

func TestAdd_DirectoryWithoutRecursiveFails(t *testing.T) {
	root, dotly := setupTest(t)

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "content")

	err := addPath(dir, root, dotly, false)
	if err == nil {
		t.Fatal("expected error for directory without -r")
	}
}

func TestAdd_DirectoryRecursive(t *testing.T) {
	root, dotly := setupTest(t)

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "a")
	writeFile(t, filepath.Join(dir, "lua/plugins/foo.lua"), "b")
	writeFile(t, filepath.Join(dir, "lazy-lock.json"), "c")

	if err := addPath(dir, root, dotly, true); err != nil {
		t.Fatalf("recursive add: %v", err)
	}

	// All three files should now be symlinks pointing into DOTLY.
	for _, rel := range []string{".config/nvim/init.lua", ".config/nvim/lua/plugins/foo.lua", ".config/nvim/lazy-lock.json"} {
		p := filepath.Join(root, rel)
		info, err := os.Lstat(p)
		if err != nil {
			t.Errorf("%s: lstat: %v", rel, err)
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s: expected symlink, got %v", rel, info.Mode())
		}
	}
}

func TestAdd_RecursiveSkipsExistingSymlinks(t *testing.T) {
	// If the tree already contains symlinks (e.g., from a partial previous add),
	// recursive add should skip them rather than erroring the whole walk.
	root, dotly := setupTest(t)

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "a")
	writeFile(t, filepath.Join(dir, "plugins.lua"), "b")

	// Pre-track one of the files.
	if err := add(filepath.Join(dir, "init.lua"), root, dotly); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Now recursive add on the parent directory should succeed.
	if err := addPath(dir, root, dotly, true); err != nil {
		t.Fatalf("recursive add over partially-tracked tree: %v", err)
	}
}
