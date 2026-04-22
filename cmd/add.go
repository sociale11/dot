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
	Use:   "add <path>...",
	Short: "Track one or more files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var failed []string
		for _, arg := range args {
			if err := addPath(arg, root, dot); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", arg, err)
				failed = append(failed, arg)
				continue
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("%d of %d failed", len(failed), len(args))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func addPath(filePath, root, dotly string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return add(absPath, root, dotly)
	}

	if info.IsDir() {
		return addDir(absPath, root, dotly)
	}

	return add(absPath, root, dotly)
}

func addDir(dirPath, root, dotly string) error {
	rel, err := filepath.Rel(root, dirPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path %s is not under root %s", dirPath, root)
	}

	dest := filepath.Join(dotly, rel)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("creating dest parent: %w", err)
	}

	if err := os.Rename(dirPath, dest); err != nil {
		return fmt.Errorf("moving directory: %w", err)
	}

	if err := os.Symlink(dest, dirPath); err != nil {
		_ = os.Rename(dest, dirPath)
		return fmt.Errorf("creating symlink: %w", err)
	}

	indexPath := filepath.Join(dotly, IndexFilename)
	if err := AddToIndex(indexPath, Index{relPath: rel, isDir: true}); err != nil {
		return fmt.Errorf("updating index: %w", err)
	}

	return nil
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
	if err := AddToIndex(indexPath, Index{relPath: rel, isDir: false}); err != nil {
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
