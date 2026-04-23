package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatus_AllHealthy(t *testing.T) {
	root, dotly := setupTest(t)

	// Add a file and a directory, both should report OK.
	file := filepath.Join(root, ".zshrc")
	writeFile(t, file, "export PATH=/foo")
	if err := add(file, root, dotly); err != nil {
		t.Fatalf("add file: %v", err)
	}

	dir := filepath.Join(root, ".config/nvim")
	writeFile(t, filepath.Join(dir, "init.lua"), "vim.opt.number = true")
	if err := addDir(dir, root, dotly); err != nil {
		t.Fatalf("add dir: %v", err)
	}

	results, err := statusCheck(root, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	for _, r := range results {
		if r.state != stateOK {
			t.Errorf("%s: expected OK, got %s", r.relPath, r.state)
		}
	}
}

func TestStatus_SymlinkMissing(t *testing.T) {
	root, dotly := setupTest(t)

	file := filepath.Join(root, ".zshrc")
	writeFile(t, file, "content")
	if err := add(file, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Simulate: someone deleted the symlink entirely.
	os.Remove(file)

	results, err := statusCheck(root, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != stateSymlinkMissing {
		t.Errorf("expected %s, got %s", stateSymlinkMissing, results[0].state)
	}
}

func TestStatus_ReplacedByRegularFile(t *testing.T) {
	root, dotly := setupTest(t)

	file := filepath.Join(root, ".zshrc")
	writeFile(t, file, "content")
	if err := add(file, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Simulate: editor safe-write replaced symlink with a regular file.
	os.Remove(file)
	writeFile(t, file, "editor wrote this")

	results, err := statusCheck(root, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != stateReplaced {
		t.Errorf("expected %s, got %s", stateReplaced, results[0].state)
	}
}

func TestStatus_SymlinkPointsWrong(t *testing.T) {
	root, dotly := setupTest(t)

	file := filepath.Join(root, ".zshrc")
	writeFile(t, file, "content")
	if err := add(file, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Simulate: symlink was replaced with one pointing elsewhere.
	os.Remove(file)
	elsewhere := filepath.Join(t.TempDir(), "wrong")
	writeFile(t, elsewhere, "wrong target")
	if err := os.Symlink(elsewhere, file); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	results, err := statusCheck(root, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != stateWrongTarget {
		t.Errorf("expected %s, got %s", stateWrongTarget, results[0].state)
	}
}

func TestStatus_SourceMissingFromRepo(t *testing.T) {
	root, dotly := setupTest(t)

	file := filepath.Join(root, ".zshrc")
	writeFile(t, file, "content")
	if err := add(file, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Simulate: someone deleted the file from the dotly repo directly.
	source := filepath.Join(dotly, ".zshrc")
	os.Remove(source)

	results, err := statusCheck(root, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != stateSourceMissing {
		t.Errorf("expected %s, got %s", stateSourceMissing, results[0].state)
	}
}

func TestStatus_EmptyIndex(t *testing.T) {
	_, dotly := setupTest(t)

	results, err := statusCheck(dotly, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestStatus_MixedHealth(t *testing.T) {
	root, dotly := setupTest(t)

	// Healthy file.
	healthy := filepath.Join(root, ".zshrc")
	writeFile(t, healthy, "zsh")
	if err := add(healthy, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Broken file: editor replaced symlink.
	broken := filepath.Join(root, ".bashrc")
	writeFile(t, broken, "bash")
	if err := add(broken, root, dotly); err != nil {
		t.Fatalf("add: %v", err)
	}
	os.Remove(broken)
	writeFile(t, broken, "editor overwrote")

	results, err := statusCheck(root, dotly)
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	states := map[string]entryState{}
	for _, r := range results {
		states[r.relPath] = r.state
	}

	if states[".zshrc"] != stateOK {
		t.Errorf(".zshrc: expected OK, got %s", states[".zshrc"])
	}
	if states[".bashrc"] != stateReplaced {
		t.Errorf(".bashrc: expected %s, got %s", stateReplaced, states[".bashrc"])
	}
}
