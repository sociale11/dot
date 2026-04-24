package cmd

import (
	"fmt"
	"os"
	"os/exec"
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

func dotInit(_, dot string) error {
	if err := os.MkdirAll(dot, 0755); err != nil {
		return fmt.Errorf("creating dot dir: %w", err)
	}

	if _, err := os.Stat(filepath.Join(dot, ".git/HEAD")); os.IsNotExist(err) {
		gitInit := exec.Command("git", "init", dot)
		if out, err := gitInit.CombinedOutput(); err != nil {
			return fmt.Errorf("git init: %w\n%s", err, out)
		}
	}

	gitignorePath := filepath.Join(dot, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		content := "**/.git\nbackups/\n"
		if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing .gitignore: %w", err)
		}
	}

	return InitIndex(filepath.Join(dot, IndexFilename))
}
