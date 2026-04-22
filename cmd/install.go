package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Recreate all symlinks from the index",
	RunE: func(cmd *cobra.Command, args []string) error {
		return install(root, dot)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func install(root, dot string) error {
	indexPath := filepath.Join(dot, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("nothing to install: index is empty")
		return nil
	}

	var failed []string
	for _, e := range entries {
		if err := installOne(e, root, dot); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", e.relPath, err)
			failed = append(failed, e.relPath)
			continue
		}
		fmt.Printf("  ✓ %s\n", e.relPath)
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d of %d entries failed", len(failed), len(entries))
	}
	return nil
}

func installOne(e Index, root, dot string) error {
	source := filepath.Join(dot, e.relPath)
	target := filepath.Join(root, e.relPath)

	// Tracked file/dir must exist in dot.
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("tracked entry missing: %w", err)
	}

	// Check what's currently at the target location.
	info, err := os.Lstat(target)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existing, err := os.Readlink(target)
			if err == nil && existing == source {
				return nil // already installed
			}
			return fmt.Errorf("symlink exists but points to %s, not %s", existing, source)
		}
		return fmt.Errorf("file already exists at %s (not a symlink)", target)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking location: %w", err)
	}

	// Make sure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("creating parent dir: %w", err)
	}

	return os.Symlink(source, target)
}
