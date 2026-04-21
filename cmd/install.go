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
		return install(root, dotly)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func install(root, dotly string) error {
	indexPath := filepath.Join(dotly, IndexFilename)
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
		if err := installOne(e); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", e.location, err)
			failed = append(failed, e.location)
			continue
		}
		fmt.Printf("  ✓ %s\n", e.location)
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d of %d entries failed", len(failed), len(entries))
	}
	return nil
}

func installOne(e Index) error {
	// Tracked file must exist in DOTLY.
	if _, err := os.Stat(e.symlink); err != nil {
		return fmt.Errorf("tracked file missing: %w", err)
	}

	// Check what's currently at the original location.
	info, err := os.Lstat(e.location)
	if err == nil {
		// Something's there. Is it already the correct symlink?
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(e.location)
			if err == nil && target == e.symlink {
				return nil // already installed, nothing to do
			}
			return fmt.Errorf("symlink exists but points elsewhere: %s", target)
		}
		return fmt.Errorf("file already exists at %s (not a symlink)", e.location)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking location: %w", err)
	}

	// Make sure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(e.location), 0755); err != nil {
		return fmt.Errorf("creating parent dir: %w", err)
	}

	return os.Symlink(e.symlink, e.location)
}
