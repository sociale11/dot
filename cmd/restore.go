package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <path>",
	Short: "Restore a tracked file to its original location and stop tracking it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return restore(args[0], root, dot)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func restore(filePath, root, dot string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	rel, err := filepath.Rel(root, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path %s is not under root %s", absPath, root)
	}

	tracked := filepath.Join(dot, rel)

	info, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("checking %s: %w", absPath, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symlink; refusing to touch it", absPath)
	}

	target, err := os.Readlink(absPath)
	if err != nil {
		return fmt.Errorf("reading symlink %s: %w", absPath, err)
	}
	if target != tracked {
		return fmt.Errorf("symlink %s points to %s, not %s; refusing to restore",
			absPath, target, tracked)
	}

	if _, err := os.Stat(tracked); err != nil {
		return fmt.Errorf("tracked file %s is missing: %w", tracked, err)
	}

	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("removing symlink: %w", err)
	}

	if err := os.Rename(tracked, absPath); err != nil {
		return fmt.Errorf("restoring file: %w", err)
	}

	indexPath := filepath.Join(dot, IndexFilename)
	if err := RemoveFromIndex(indexPath, rel); err != nil {
		return fmt.Errorf("updating index: %w", err)
	}

	return nil
}
