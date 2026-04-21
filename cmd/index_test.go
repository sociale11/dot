package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func indexPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "dotly.index")
}

func TestReadIndex_MissingFile(t *testing.T) {
	// Reading a non-existent index should return an empty slice, not an error.
	// (Or an error, depending on how you designed it — adjust the assertion.)
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
		{location: "/home/x/.zshrc", symlink: "/home/x/.local/share/dotly/.zshrc"},
		{location: "/home/x/.config/nvim/init.lua", symlink: "/home/x/.local/share/dotly/.config/nvim/init.lua"},
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
	// Writing a second time should replace the file, not append.
	path := indexPath(t)

	first := []Index{{location: "/a", symlink: "/b"}}
	second := []Index{{location: "/c", symlink: "/d"}}

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

	// Manually write a file with blank lines mixed in.
	content := "/a\t/b\n\n/c\t/d\n\n"
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
	// Lines without a tab separator should be skipped, not crash.
	path := indexPath(t)

	content := "/a\t/b\nmalformed-line-no-tab\n/c\t/d\n"
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

	if err := AddToIndex(path, Index{location: "/a", symlink: "/b"}); err != nil {
		t.Fatalf("AddToIndex: %v", err)
	}
	if err := AddToIndex(path, Index{location: "/c", symlink: "/d"}); err != nil {
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
	// Adding the same location twice should not create a duplicate entry.
	path := indexPath(t)

	entry := Index{location: "/a", symlink: "/b"}
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

func TestAddToIndex_ReplacesOnSameLocation(t *testing.T) {
	// Adding the same location with a different symlink should replace, not duplicate.
	path := indexPath(t)

	if err := AddToIndex(path, Index{location: "/a", symlink: "/old"}); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := AddToIndex(path, Index{location: "/a", symlink: "/new"}); err != nil {
		t.Fatalf("second add: %v", err)
	}

	got, err := ReadIndex(path)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].symlink != "/new" {
		t.Errorf("expected symlink /new, got %s", got[0].symlink)
	}
}
