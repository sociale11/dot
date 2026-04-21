package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Track a file by moving it to dotly and symlinking it back",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return add(args[0], root, dotly)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func add(filePath, root, dotly string) error {
	if err := os.MkdirAll(dotly, 0755); err != nil {
		return fmt.Errorf("initializing dotly dir: %w", err)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	if info, err := os.Lstat(absPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(absPath)
			if err != nil {
				return fmt.Errorf("inspecting existing symlink: %w", err)
			}
			// Is it a symlink into our DOTLY?
			if rel, err := filepath.Rel(dotly, target); err == nil && !strings.HasPrefix(rel, "..") {
				return fmt.Errorf("%s is already tracked (symlink to %s)", absPath, target)
			}
			return fmt.Errorf("%s is a symlink to %s (not managed by dotly); refusing to touch it", absPath, target)
		}
	}

	rel, err := filepath.Rel(root, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path %s is not under root %s", absPath, root)
	}

	dest := filepath.Join(dotly, rel)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("creating dest dir: %w", err)
	}
	if err := copyFile(absPath, dest); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing original: %w", err)
	}
	if err := os.Symlink(dest, absPath); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	indexPath := filepath.Join(dotly, IndexFilename)
	if err := AddToIndex(indexPath, Index{location: absPath, symlink: dest}); err != nil {
		return fmt.Errorf("updating index: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("copying: %w", err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("finalizing destination: %w", err)
	}
	return nil
}
