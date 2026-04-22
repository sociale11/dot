package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func indexPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "dot.index")
}

func TestReadIndex_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist")

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("expected no error for missing index, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestWriteIndex_ThenReadIndex(t *testing.T) {
	path := indexPath(t)

	want := []Index{
		{relPath: ".zshrc", isDir: false},
		{relPath: ".config/nvim", isDir: true},
		{relPath: ".config/git/config", isDir: false},
	}

	if err := WriteIndex(path, want); err != nil {
		t.Fatalf("WriteIndex: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("length: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestWriteIndex_FileFormat(t *testing.T) {
	// Verify the actual on-disk format is what we expect.
	path := indexPath(t)

	entries := []Index{
		{relPath: ".zshrc", isDir: false},
		{relPath: ".config/nvim", isDir: true},
	}

	if err := WriteIndex(path, entries); err != nil {
		t.Fatalf("WriteIndex: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	expected := ".zshrc\tfile\n.config/nvim\tdir\n"
	if string(raw) != expected {
		t.Errorf("file content:\ngot:  %q\nwant: %q", string(raw), expected)
	}
}

func TestWriteIndex_EmptySlice(t *testing.T) {
	path := indexPath(t)

	if err := WriteIndex(path, []Index{}); err != nil {
		t.Fatalf("WriteIndex: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestWriteIndex_Overwrites(t *testing.T) {
	path := indexPath(t)

	first := []Index{{relPath: ".zshrc", isDir: false}}
	second := []Index{{relPath: ".bashrc", isDir: false}}

	if err := WriteIndex(path, first); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteIndex(path, second); err != nil {
		t.Fatalf("second write: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 1 || got[0] != second[0] {
		t.Errorf("expected %v, got %v", second, got)
	}
}

func TestReadIndex_SkipsBlankLines(t *testing.T) {
	path := indexPath(t)

	content := ".zshrc\tfile\n\n.config/nvim\tdir\n\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(got), got)
	}
}

func TestReadIndex_SkipsMalformedLines(t *testing.T) {
	path := indexPath(t)

	content := ".zshrc\tfile\nmalformed-no-tab\n.config/nvim\tdir\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 valid entries, got %d: %v", len(got), got)
	}
}

func TestReadIndex_SkipsInvalidKind(t *testing.T) {
	path := indexPath(t)

	content := ".zshrc\tfile\n.bashrc\tgarbage\n.config/nvim\tdir\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 valid entries, got %d: %v", len(got), got)
	}
}

func TestAddToIndex_AppendsNew(t *testing.T) {
	path := indexPath(t)

	if err := AddToIndex(path, Index{relPath: ".zshrc", isDir: false}); err != nil {
		t.Fatalf("AddToIndex: %v", err)
	}
	if err := AddToIndex(path, Index{relPath: ".config/nvim", isDir: true}); err != nil {
		t.Fatalf("AddToIndex: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestAddToIndex_IsIdempotent(t *testing.T) {
	path := indexPath(t)

	entry := Index{relPath: ".zshrc", isDir: false}
	if err := AddToIndex(path, entry); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := AddToIndex(path, entry); err != nil {
		t.Fatalf("second add: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry after duplicate add, got %d: %v", len(got), got)
	}
}

func TestAddToIndex_ReplacesOnSamePath(t *testing.T) {
	// If something was tracked as file then re-added as dir (unlikely but
	// defensively correct), the kind should update.
	path := indexPath(t)

	if err := AddToIndex(path, Index{relPath: ".config/nvim", isDir: false}); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := AddToIndex(path, Index{relPath: ".config/nvim", isDir: true}); err != nil {
		t.Fatalf("second add: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if !got[0].isDir {
		t.Errorf("expected isDir=true after replacement, got false")
	}
}

func TestRemoveFromIndex(t *testing.T) {
	path := indexPath(t)

	if err := AddToIndex(path, Index{relPath: ".zshrc", isDir: false}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := AddToIndex(path, Index{relPath: ".config/nvim", isDir: true}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := RemoveFromIndex(path, ".zshrc"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].relPath != ".config/nvim" {
		t.Errorf("wrong entry survived: got %s", got[0].relPath)
	}
}

func TestRemoveFromIndex_AbsentIsNoOp(t *testing.T) {
	path := indexPath(t)

	if err := AddToIndex(path, Index{relPath: ".zshrc", isDir: false}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := RemoveFromIndex(path, ".nonexistent"); err != nil {
		t.Fatalf("remove absent: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry unchanged, got %d", len(got))
	}
}
