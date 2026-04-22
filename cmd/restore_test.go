package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestore_FileRoundTrip(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "original content")

	if err := add(original, root, dot); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := restore(original, root, dot); err != nil {
		t.Fatalf("restore: %v", err)
	}

	// Should be a regular file again.
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected regular file, got symlink")
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
	tracked := filepath.Join(dot, ".zshrc")
	if _, err := os.Stat(tracked); !os.IsNotExist(err) {
		t.Errorf("tracked file should be removed")
	}

	// Index should be empty.
	indexPath := filepath.Join(dot, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty index, got %d entries", len(entries))
	}
}

func TestRestore_DirectoryRoundTrip(t *testing.T) {
	root, dot := setupTest(t)

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "a")
	writeFile(t, filepath.Join(dir, "lua/plugins.lua"), "b")

	if err := addDir(dir, root, dot); err != nil {
		t.Fatalf("add dir: %v", err)
	}
	if err := restore(dir, root, dot); err != nil {
		t.Fatalf("restore dir: %v", err)
	}

	// Should be a real directory again.
	info, err := os.Lstat(dir)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected regular directory, got symlink")
	}
	if !info.IsDir() {
		t.Errorf("expected directory, got %v", info.Mode())
	}

	// Files inside should still exist.
	got, err := os.ReadFile(filepath.Join(dir, "init.lua"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "a" {
		t.Errorf("content: got %q, want %q", got, "a")
	}

	gotNested, err := os.ReadFile(filepath.Join(dir, "lua/plugins.lua"))
	if err != nil {
		t.Fatalf("read nested: %v", err)
	}
	if string(gotNested) != "b" {
		t.Errorf("nested content: got %q, want %q", gotNested, "b")
	}

	// Tracked copy gone from dot repo.
	tracked := filepath.Join(dot, ".config/nvim")
	if _, err := os.Stat(tracked); !os.IsNotExist(err) {
		t.Errorf("tracked dir should be removed from repo")
	}
}

func TestRestore_NestedFile(t *testing.T) {
	root, dot := setupTest(t)

	original := filepath.Join(root, ".config/git/config")
	writeFile(t, original, "[user]\n\tname = test")

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
		t.Errorf("expected regular file, got symlink")
	}

	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "[user]\n\tname = test" {
		t.Errorf("content: got %q", got)
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
		t.Errorf("error: got %q, want mention of 'not a symlink'", err)
	}
}

func TestRestore_RejectsForeignSymlink(t *testing.T) {
	root, dot := setupTest(t)

	elsewhere := filepath.Join(t.TempDir(), "other")
	writeFile(t, elsewhere, "not ours")

	original := filepath.Join(root, ".zshrc")
	if err := os.Symlink(elsewhere, original); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	err := restore(original, root, dot)
	if err == nil {
		t.Fatal("expected error for foreign symlink, got nil")
	}

	// Symlink should be untouched.
	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != elsewhere {
		t.Errorf("symlink was modified: got %s, want %s", target, elsewhere)
	}
}

func TestRestore_RejectsPathOutsideRoot(t *testing.T) {
	root, dot := setupTest(t)

	outside := filepath.Join(t.TempDir(), "evil")
	writeFile(t, outside, "nope")

	err := restore(outside, root, dot)
	if err == nil {
		t.Fatal("expected error for path outside root, got nil")
	}
	if !strings.Contains(err.Error(), "not under root") {
		t.Errorf("error: got %q, want mention of 'not under root'", err)
	}
}

func TestRestore_RejectsWhenTrackedFileMissing(t *testing.T) {
	root, dot := setupTest(t)

	file := filepath.Join(root, ".zshrc")
	writeFile(t, file, "content")

	if err := add(file, root, dot); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Delete the file from the repo but leave the symlink.
	tracked := filepath.Join(dot, ".zshrc")
	os.Remove(tracked)

	err := restore(file, root, dot)
	if err == nil {
		t.Fatal("expected error when tracked file is missing, got nil")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error: got %q, want mention of 'missing'", err)
	}

	// Symlink should still exist (restore didn't partially execute).
	info, err := os.Lstat(file)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("symlink should still exist after failed restore")
	}
}

func TestRestore_OnlyRemovesItsOwnIndexEntry(t *testing.T) {
	root, dot := setupTest(t)

	fileA := filepath.Join(root, ".zshrc")
	fileB := filepath.Join(root, ".bashrc")
	writeFile(t, fileA, "zsh")
	writeFile(t, fileB, "bash")

	if err := add(fileA, root, dot); err != nil {
		t.Fatalf("add A: %v", err)
	}
	if err := add(fileB, root, dot); err != nil {
		t.Fatalf("add B: %v", err)
	}

	if err := restore(fileA, root, dot); err != nil {
		t.Fatalf("restore A: %v", err)
	}

	// .bashrc should still be in the index.
	indexPath := filepath.Join(dot, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].relPath != ".bashrc" {
		t.Errorf("wrong entry survived: got %s, want .bashrc", entries[0].relPath)
	}
}
