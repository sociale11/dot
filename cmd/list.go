package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all tracked entries from the index",
	RunE: func(cmd *cobra.Command, args []string) error {
		return list(root, dot)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func list(root, dot string) error {
	indexPath := filepath.Join(dot, IndexFilename)
	entries, err := ReadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}

	for _, e := range entries {
		fmt.Printf("%s\t%t\n", e.relPath, e.isDir)
	}
	return nil
}
