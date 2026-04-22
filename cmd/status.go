package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show tracked entries and their health",
	RunE: func(cmd *cobra.Command, args []string) error {
		return status(root, dot)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

type entryState string

const (
	stateOK             entryState = "ok"
	stateSymlinkMissing entryState = "symlink_missing"
	stateReplaced       entryState = "replaced"
	stateWrongTarget    entryState = "wrong_target"
	stateSourceMissing  entryState = "source_missing"
)

type statusResult struct {
	relPath string
	state   entryState
}

func status(root, dotly string) error {
	results, err := statusCheck(root, dotly)
	if err != nil {
		return err
	}
	for _, r := range results {
		if r.state == stateOK {
			fmt.Printf("  ✓ %s\n", r.relPath)
		} else {
			fmt.Printf("  ✗ %s: %s\n", r.relPath, r.state)
		}
	}
	return nil
}

func statusCheck(root, dotly string) ([]statusResult, error) {
	indexPath := filepath.Join(dotly, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		return nil, err
	}

	var results []statusResult
	for _, e := range entries {
		source := filepath.Join(dotly, e.relPath)
		target := filepath.Join(root, e.relPath)

		if _, err := os.Stat(source); err != nil {
			results = append(results, statusResult{e.relPath, stateSourceMissing})
			continue
		}

		info, err := os.Lstat(target)
		if err != nil {
			results = append(results, statusResult{e.relPath, stateSymlinkMissing})
			continue
		}

		if info.Mode()&os.ModeSymlink == 0 {
			results = append(results, statusResult{e.relPath, stateReplaced})
			continue
		}

		link, err := os.Readlink(target)
		if err != nil || link != source {
			results = append(results, statusResult{e.relPath, stateWrongTarget})
			continue
		}

		results = append(results, statusResult{e.relPath, stateOK})
	}
	return results, nil
}
