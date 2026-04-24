package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupBareRepo creates a bare git repo with some dotfiles, returns the repo path.
func setupBareRepo(t *testing.T, root string) string {
	t.Helper()

	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	run("git", "init")

	writeFile(t, filepath.Join(work, ".zshrc"), "export PATH=/foo")

	idx := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(filepath.Join(work, IndexFilename), idx); err != nil {
		t.Fatalf("seed index: %v", err)
	}

	run("git", "add", ".")
	run("git", "commit", "-m", "initial")

	bare := filepath.Join(t.TempDir(), "dotfiles.git")
	run2 := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := run2.CombinedOutput(); err != nil {
		t.Fatalf("bare clone: %v\n%s", err, out)
	}

	return bare
}

func TestClone_ClonesAndInstalls(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t, root)

	if err := cloneAndInstall(bare, root, dot, false); err != nil {
		t.Fatalf("cloneAndInstall: %v", err)
	}

	// Index should exist in dot.
	indexPath := filepath.Join(dot, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 index entry, got %d", len(entries))
	}

	// Symlink should exist at root.
	original := filepath.Join(root, ".zshrc")
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink at original location")
	}

	// Content readable through symlink.
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "export PATH=/foo" {
		t.Errorf("content: got %q, want %q", got, "export PATH=/foo")
	}
}

func TestClone_RefusesNonEmptyDotDir(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t, root)

	// Pre-populate dot directory.
	writeFile(t, filepath.Join(dot, "something"), "existing stuff")

	err := cloneAndInstall(bare, root, dot, false)
	if err == nil {
		t.Fatal("expected error when dot dir is non-empty")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Errorf("error: got %q, want mention of 'not empty'", err)
	}
}

func TestClone_OverwriteConflicts(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t, root)

	// Create a conflicting file at the target location.
	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "existing config")

	if err := cloneAndInstall(bare, root, dot, true); err != nil {
		t.Fatalf("cloneAndInstall --overwrite: %v", err)
	}

	// Original should now be a symlink.
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink after overwrite")
	}

	// Backup should exist.
	backup := filepath.Join(dot, "backups", ".zshrc")
	got, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != "existing config" {
		t.Errorf("backup content: got %q, want %q", got, "existing config")
	}
}

func TestClone_FailsOnConflictWithoutOverwrite(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t, root)

	// Create a conflicting file.
	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "do not touch")

	err := cloneAndInstall(bare, root, dot, false)
	if err == nil {
		t.Fatal("expected error on conflict without overwrite")
	}

	// User's file must be untouched.
	got, _ := os.ReadFile(original)
	if string(got) != "do not touch" {
		t.Errorf("file was modified: got %q", got)
	}
}

func TestClone_BadUrlFails(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	err := cloneAndInstall("https://example.com/nonexistent/repo.git", root, dot, false)
	if err == nil {
		t.Fatal("expected error for bad repo URL")
	}
	if !strings.Contains(err.Error(), "git clone failed") {
		t.Errorf("error: got %q, want mention of 'git clone failed'", err)
	}
}
