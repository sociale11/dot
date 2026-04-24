package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_CreatesDirectory(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("dotInit: %v", err)
	}

	info, err := os.Stat(dot)
	if err != nil {
		t.Fatalf("dot dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestInit_InitializesGitRepo(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("dotInit: %v", err)
	}

	head := filepath.Join(dot, ".git/HEAD")
	if _, err := os.Stat(head); err != nil {
		t.Errorf("expected .git/HEAD to exist (git init didn't run): %v", err)
	}
}

func TestInit_CreatesGitignore(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("dotInit: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dot, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}

	if !strings.Contains(string(content), "**/.git") {
		t.Error(".gitignore missing **/.git entry")
	}
	if !strings.Contains(string(content), "backups/") {
		t.Error(".gitignore missing backups/ entry")
	}
}

func TestInit_CreatesIndexFile(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("dotInit: %v", err)
	}

	indexPath := filepath.Join(dot, IndexFilename)
	if _, err := os.Stat(indexPath); err != nil {
		t.Errorf("index file not created: %v", err)
	}
}

func TestInit_IdempotentOnExistingRepo(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("first init: %v", err)
	}

	writeFile(t, filepath.Join(dot, ".zshrc"), "content")
	if err := AddToIndex(filepath.Join(dot, IndexFilename), Index{relPath: ".zshrc", isDir: false}); err != nil {
		t.Fatalf("AddToIndex: %v", err)
	}

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("second init: %v", err)
	}

	entries, err := ReadIndex(filepath.Join(dot, IndexFilename))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after re-init, got %d", len(entries))
	}
}

func TestInit_DoesNotOverwriteExistingGitignore(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("first init: %v", err)
	}

	gitignorePath := filepath.Join(dot, ".gitignore")
	existing, _ := os.ReadFile(gitignorePath)
	if err := os.WriteFile(gitignorePath, append(existing, []byte("\n*.secret\n")...), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := dotInit(root, dot); err != nil {
		t.Fatalf("second init: %v", err)
	}

	content, _ := os.ReadFile(gitignorePath)
	if !strings.Contains(string(content), "*.secret") {
		t.Error("re-init overwrote user's .gitignore additions")
	}
}
