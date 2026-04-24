package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var cloneOverwrite bool

var cloneCmd = &cobra.Command{
	Use:   "clone <repo-url>",
	Short: "Clone a dotfiles repo and install symlinks",
	Long: `Clone a dotfiles repo into the dot storage directory and run install.

This is the recommended way to set up dot on a new machine:
  dot clone git@github.com:user/dotfiles.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cloneAndInstall(args[0], root, dot, cloneOverwrite)
	},
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneOverwrite, "overwrite", false, "backup conflicting files and replace them")
	rootCmd.AddCommand(cloneCmd)
}

func cloneAndInstall(url, root, dot string, overwrite bool) error {
	// Refuse if dot directory already exists and is non-empty.
	entries, err := os.ReadDir(dot)
	if err == nil && len(entries) > 0 {
		return fmt.Errorf("%s already exists and is not empty; remove it first or use 'dot install'", dot)
	}

	fmt.Printf("Cloning %s into %s...\n", url, dot)
	gitClone := exec.Command("git", "clone", url, dot)
	gitClone.Stdout = os.Stdout
	gitClone.Stderr = os.Stderr
	if err := gitClone.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Println("Installing symlinks...")
	return install(root, dot, overwrite)
}
