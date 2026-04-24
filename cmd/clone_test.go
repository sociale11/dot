package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupBareRepo creates a bare git repo with a .zshrc and index on main.
func setupBareRepo(t *testing.T) string {
	t.Helper()

	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
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
	if err := AddToIndex(filepath.Join(work, IndexFilename), Index{relPath: ".zshrc", isDir: false}); err != nil {
		t.Fatalf("AddToIndex: %v", err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "initial")

	bare := filepath.Join(t.TempDir(), "dotfiles.git")
	cmd := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bare clone: %v\n%s", err, out)
	}

	return bare
}

// setupBareRepoWithBranch creates a bare repo with main and a named branch
// that has different content.
func setupBareRepoWithBranch(t *testing.T, branch string) string {
	t.Helper()

	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
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

	// main branch
	run("git", "init", "-b", "main")
	writeFile(t, filepath.Join(work, ".zshrc"), "main content")
	if err := AddToIndex(filepath.Join(work, IndexFilename), Index{relPath: ".zshrc", isDir: false}); err != nil {
		t.Fatalf("AddToIndex: %v", err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "initial")

	// create named branch with different content
	run("git", "checkout", "-b", branch)
	writeFile(t, filepath.Join(work, ".zshrc"), "branch content")
	run("git", "add", ".")
	run("git", "commit", "-m", "branch commit")
	run("git", "checkout", "main")

	bare := filepath.Join(t.TempDir(), "dotfiles.git")
	cmd := exec.Command("git", "clone", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bare clone: %v\n%s", err, out)
	}

	return bare
}

func TestClone_ClonesAndInstalls(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t)

	if err := cloneAndInstall([]string{bare}, root, dot); err != nil {
		t.Fatalf("cloneAndInstall: %v", err)
	}

	// Symlink should exist.
	original := filepath.Join(root, ".zshrc")
	info, err := os.Lstat(original)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink")
	}

	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "export PATH=/foo" {
		t.Errorf("content: got %q, want %q", got, "export PATH=/foo")
	}
}

func TestClone_WithBranchFlag(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepoWithBranch(t, "desktop")

	if err := cloneAndInstall([]string{"-b", "desktop", bare}, root, dot); err != nil {
		t.Fatalf("cloneAndInstall -b desktop: %v", err)
	}

	// Should have the branch content, not main.
	original := filepath.Join(root, ".zshrc")
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "branch content" {
		t.Errorf("content: got %q, want %q (should be branch, not main)", got, "branch content")
	}

	// Verify we're on the right branch.
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dot
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git branch: %v", err)
	}
	if strings.TrimSpace(string(out)) != "desktop" {
		t.Errorf("branch: got %q, want %q", strings.TrimSpace(string(out)), "desktop")
	}
}

func TestClone_RefusesNonEmptyDotDir(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t)

	writeFile(t, filepath.Join(dot, "something"), "existing")

	err := cloneAndInstall([]string{bare}, root, dot)
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
	bare := setupBareRepo(t)

	// Create conflicting file.
	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "existing config")

	overwrite = true
	defer func() { overwrite = false }()

	if err := cloneAndInstall([]string{bare}, root, dot); err != nil {
		t.Fatalf("cloneAndInstall --overwrite: %v", err)
	}

	// Should be a symlink now.
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
		t.Errorf("backup: got %q, want %q", got, "existing config")
	}
}

func TestClone_FailsOnConflictWithoutOverwrite(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t)

	original := filepath.Join(root, ".zshrc")
	writeFile(t, original, "do not touch")

	overwrite = false

	err := cloneAndInstall([]string{bare}, root, dot)
	if err == nil {
		t.Fatal("expected error on conflict without overwrite")
	}

	got, _ := os.ReadFile(original)
	if string(got) != "do not touch" {
		t.Errorf("file was modified: got %q", got)
	}
}

func TestClone_BadUrlFails(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	err := cloneAndInstall([]string{"/nonexistent/repo.git"}, root, dot)
	if err == nil {
		t.Fatal("expected error for bad repo")
	}
	if !strings.Contains(err.Error(), "git clone failed") {
		t.Errorf("error: got %q, want mention of 'git clone failed'", err)
	}
}

func TestClone_NoArgsFails(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")

	err := cloneAndInstall([]string{}, root, dot)
	if err == nil {
		t.Fatal("expected error with no args")
	}
}

func TestClone_WithDepthFlag(t *testing.T) {
	root := t.TempDir()
	dot := filepath.Join(root, ".local/share/dot")
	bare := setupBareRepo(t)

	// Should pass --depth through to git.
	if err := cloneAndInstall([]string{"--depth", "1", bare}, root, dot); err != nil {
		t.Fatalf("cloneAndInstall --depth 1: %v", err)
	}

	// Verify shallow clone.
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = dot
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-list: %v", err)
	}
	if strings.TrimSpace(string(out)) != "1" {
		t.Errorf("expected shallow clone with 1 commit, got %s", strings.TrimSpace(string(out)))
	}
}
