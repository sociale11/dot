package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var overwrite bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Recreate all symlinks from the index",
	RunE: func(cmd *cobra.Command, args []string) error {
		return install(root, dot, overwrite)
	},
}

func init() {
	installCmd.Flags().BoolVar(&overwrite, "overwrite", false, "backup conflicting files and replace them")
	rootCmd.AddCommand(installCmd)
}

func install(root, dot string, overwrite bool) error {
	indexPath := filepath.Join(dot, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("nothing to install: index is empty")
		return nil
	}

	var errors []string
	for _, e := range entries {
		if err := installOne(e, root, dot, overwrite); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", e.relPath, err)
			errors = append(errors, fmt.Sprintf("%s: %v", e.relPath, err))
			continue
		}
		fmt.Printf("  ✓ %s\n", e.relPath)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d of %d entries failed:\n%s", len(errors), len(entries), strings.Join(errors, "\n"))
	}
	return nil
}

func installOne(e Index, root, dot string, overwrite bool) error {
	source := filepath.Join(dot, e.relPath)
	target := filepath.Join(root, e.relPath)

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("tracked entry missing: %w", err)
	}

	info, err := os.Lstat(target)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existing, err := os.Readlink(target)
			if err == nil && existing == source {
				return nil
			}
		}

		// Something exists at the target and it's not the right symlink.
		if !overwrite {
			return fmt.Errorf("conflict at %s (use --overwrite to backup and replace)", target)
		}

		if err := backup(dot, e, root); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking location: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("creating parent dir: %w", err)
	}

	return os.Symlink(source, target)
}

func backup(dot string, e Index, root string) error {
	source := filepath.Join(root, e.relPath)
	dest := filepath.Join(dot, "backups", e.relPath)

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("creating backup dir: %w", err)
	}

	if err := os.Rename(source, dest); err != nil {
		return fmt.Errorf("backing up %s: %w", e.relPath, err)
	}

	fmt.Printf("  ⟳ backed up %s\n", e.relPath)
	return nil
}
