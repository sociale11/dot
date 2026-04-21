package cmd

import (
	"fmt"
	"os"
	"strings"
)

type Index struct {
	location string // location of the tracked file
	symlink  string // location of the symlink
}

const IndexFilename = "dotly.index"

// InitIndex creates an empty index file at the given path if it doesn't exist.
// Safe to call repeatedly.
func InitIndex(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("initializing index at %s: %w", path, err)
	}
	return f.Close()
}

// WriteIndex replaces the index file at path with the given entries.
func WriteIndex(path string, indexes []Index) error {
	var sb strings.Builder
	for _, idx := range indexes {
		fmt.Fprintf(&sb, "%s\t%s\n", idx.location, idx.symlink)
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing index to %s: %w", path, err)
	}
	return nil
}

// ReadIndex loads the index from path. A missing file returns an empty slice
// and no error — that's a normal "nothing tracked yet" state. Malformed lines
// are skipped silently.
func ReadIndex(path string) ([]Index, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Index{}, nil
		}
		return nil, fmt.Errorf("reading index at %s: %w", path, err)
	}

	var indexes []Index
	for _, line := range strings.Split(string(content), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue
		}
		indexes = append(indexes, Index{
			location: parts[0],
			symlink:  parts[1],
		})
	}
	return indexes, nil
}

// AddToIndex adds or replaces an entry in the index. If an entry with the
// same location already exists, it's updated in place. Otherwise the new
// entry is appended.
func AddToIndex(path string, idx Index) error {
	existing, err := ReadIndex(path)
	if err != nil {
		return err
	}

	found := false
	for i, e := range existing {
		if e.location == idx.location {
			existing[i] = idx
			found = true
			break
		}
	}
	if !found {
		existing = append(existing, idx)
	}

	return WriteIndex(path, existing)
}

// RemoveFromIndex removes the entry for the given location. Returns nil if
// the location wasn't in the index — removing something already absent is
// not an error.
func RemoveFromIndex(path, location string) error {
	existing, err := ReadIndex(path)
	if err != nil {
		return err
	}

	filtered := existing[:0] // reuse underlying array
	for _, e := range existing {
		if e.location != location {
			filtered = append(filtered, e)
		}
	}
	return WriteIndex(path, filtered)
}
