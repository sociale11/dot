package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstall_RecreatesSymlinks(t *testing.T) {
	root, dot := setupTest(t)

	// Simulate a cloned dot: tracked file exists, index has the entry,
	// but the symlink at the original location is missing.
	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "export PATH=/foo")

	original := filepath.Join(root, ".zshrc")
	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed index: %v", err)
	}

	if err := install(root, dot); err != nil {
		t.Fatalf("install: %v", err)
	}

	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink at original location")
	}
}

func TestInstall_RefusesToOverwriteRegularFile(t *testing.T) {
	root, dot := setupTest(t)

	tracked := filepath.Join(dot, ".zshrc")
	writeFile(t, tracked, "tracked content")

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "user's real file, do not touch")

	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(dot, IndexFilename), idx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	err := install(root, dot)
	if err == nil {
		t.Fatal("expected install to fail when real file blocks symlink")
	}

	// User's file must be untouched.
	got, _ := os.ReadFile(original)
	if string(got) != "user's real file, do not touch" {
		t.Errorf("user's file was modified: %s", got)
	}
}
