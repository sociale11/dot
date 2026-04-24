package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall_RecreatesSymlinks(t *testing.T) {
	root, dot := setupTest(t)

	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "export PATH=/foo")

	original := filepath.Join(root, ".zshrc")
	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed index: %v", err)
	}

	if err := install(root, dot, false); err != nil {
		t.Fatalf("install: %v", err)
	}

	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink at original location")
	}

	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != tracked {
		t.Errorf("symlink target: got %s, want %s", target, tracked)
	}

	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "export PATH=/foo" {
		t.Errorf("content: got %q, want %q", got, "export PATH=/foo")
	}
}

func TestInstall_RecreatesDirectorySymlink(t *testing.T) {
	root, dot := setupTest(t)

	dir := filepath.Join(dot, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "a")
	writeFile(t, filepath.Join(dir, "lua/plugins.lua"), "b")

	idx := Index{relPath: ".config/nvim", isDir: true}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed index: %v", err)
	}

	if err := install(root, dot, false); err != nil {
		t.Fatalf("install: %v", err)
	}

	target := filepath.Join(root, ".config/nvim")
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink")
	}

	got, err := os.ReadFile(filepath.Join(target, "init.lua"))
	if err != nil {
		t.Fatalf("read through symlink: %v", err)
	}
	if string(got) != "a" {
		t.Errorf("content: got %q, want %q", got, "a")
	}
}

func TestInstall_SkipsAlreadyCorrectSymlink(t *testing.T) {
	root, dot := setupTest(t)

	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "content")

	original := filepath.Join(root, ".zshrc")
	if err := os.MkdirAll(filepath.Dir(original), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Symlink(tracked, original); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Should succeed silently — already installed.
	if err := install(root, dot, false); err != nil {
		t.Fatalf("install: %v", err)
	}
}

func TestInstall_RefusesConflictWithoutOverwrite(t *testing.T) {
	root, dot := setupTest(t)

	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "tracked content")

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "user's real file, do not touch")

	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	err := install(root, dot, false)
	if err == nil {
		t.Fatal("expected install to fail on conflict without --overwrite")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("error: got %q, want mention of 'conflict'", err)
	}

	// User's file must be untouched.
	got, _ := os.ReadFile(original)
	if string(got) != "user's real file, do not touch" {
		t.Errorf("user's file was modified: %s", got)
	}
}

func TestInstall_OverwriteBacksUpAndReplaces(t *testing.T) {
	root, dot := setupTest(t)

	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "tracked content")

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "existing content")

	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := install(root, dot, true); err != nil {
		t.Fatalf("install --overwrite: %v", err)
	}

	// Original should now be a symlink.
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink at original location")
	}

	// Backup should exist with the old content.
	backupPath := filepath.Join(dot, "backups", ".zshrc")
	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != "existing content" {
		t.Errorf("backup content: got %q, want %q", got, "existing content")
	}
}

func TestInstall_OverwriteBacksUpDirectory(t *testing.T) {
	root, dot := setupTest(t)

	// Tracked directory in repo.
	trackedDir := filepath.Join(dot, ".config/nvim")
	writeFile(t, filepath.Join(trackedDir, "init.lua"), "tracked")

	// Existing directory at target.
	existingDir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(existingDir, "init.lua"), "existing")
	writeFile(t, filepath.Join(existingDir, "local.lua"), "local only")

	idx := Index{relPath: ".config/nvim", isDir: true}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := install(root, dot, true); err != nil {
		t.Fatalf("install --overwrite: %v", err)
	}

	// Original should be a symlink.
	info, err := os.Lstat(existingDir)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink")
	}

	// Backup should contain the old directory.
	backupInit := filepath.Join(dot, "backups", ".config/nvim/init.lua")
	got, err := os.ReadFile(backupInit)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != "existing" {
		t.Errorf("backup content: got %q, want %q", got, "existing")
	}

	backupLocal := filepath.Join(dot, "backups", ".config/nvim/local.lua")
	got, err = os.ReadFile(backupLocal)
	if err != nil {
		t.Fatalf("read backup local: %v", err)
	}
	if string(got) != "local only" {
		t.Errorf("backup local content: got %q, want %q", got, "local only")
	}
}

func TestInstall_OverwriteWrongSymlink(t *testing.T) {
	root, dot := setupTest(t)

	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "tracked")

	// Symlink exists but points somewhere else.
	original := filepath.Join(root, ".zshrc")
	elsewhere := filepath.Join(t.TempDir(), "wrong")
	writeFile(t, elsewhere, "wrong target")
	if err := os.MkdirAll(filepath.Dir(original), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Symlink(elsewhere, original); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := install(root, dot, true); err != nil {
		t.Fatalf("install --overwrite: %v", err)
	}

	// Should now point to tracked.
	target, err := os.Readlink(original)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != tracked {
		t.Errorf("symlink target: got %s, want %s", target, tracked)
	}
}

func TestInstall_EmptyIndex(t *testing.T) {
	root, dot := setupTest(t)

	if err := install(root, dot, false); err != nil {
		t.Fatalf("install empty: %v", err)
	}
}

func TestInstall_MissingTrackedEntry(t *testing.T) {
	root, dot := setupTest(t)

	// Index references a file that doesn't exist in the repo.
	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	err := install(root, dot, false)
	if err == nil {
		t.Fatal("expected error for missing tracked entry")
	}
}

func TestInstall_MultipleEntries(t *testing.T) {
	root, dot := setupTest(t)

	writeFile(t, filepath.Join(dot, ".zshrc"), "zsh")
	writeFile(t, filepath.Join(dot, ".gitconfig"), "git")
	writeFile(t, filepath.Join(dot, ".config/nvim/init.lua"), "nvim")

	indexPath := filepath.Join(dot, IndexFilename)
	for _, idx := range []Index{
		{relPath: ".zshrc", isDir: false},
		{relPath: ".gitconfig", isDir: false},
		{relPath: ".config/nvim", isDir: true},
	} {
		if err := AddToIndex(indexPath, idx); err != nil {
			t.Fatalf("AddToIndex: %v", err)
		}
	}

	if err := install(root, dot, false); err != nil {
		t.Fatalf("install: %v", err)
	}

	// All three should be symlinks.
	for _, rel := range []string{".zshrc", ".gitconfig", ".config/nvim"} {
		p := filepath.Join(root, rel)
		info, err := os.Lstat(p)
		if err != nil {
			t.Errorf("%s: lstat: %v", rel, err)
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s: expected symlink", rel)
		}
	}
}
