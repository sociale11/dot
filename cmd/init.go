package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the dot folder, initialize git and creates .gitignore",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dotInit(root, dot)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func dotInit(root, dot string) error {
	if err := os.MkdirAll(dot, 0755); err != nil {
		return fmt.Errorf("creating dot dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(dot, ".git"), 0755); err != nil {
		return fmt.Errorf("creating git dir: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dot, ".gitignore"), nil, 0644); err != nil {
		return fmt.Errorf("writing .gitignore: %w", err)
	}

	return nil
}
