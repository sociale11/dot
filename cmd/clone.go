package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var cloneOverwrite bool

var cloneCmd = &cobra.Command{
	Use:   "clone [flags] <repo-url>",
	Short: "Clone a dotfiles repo and install symlinks",
	Long: `Clone a dotfiles repo into the dot storage directory and run install.

Examples:
  dot clone git@github.com:user/dotfiles.git
  dot clone -b desktop git@github.com:user/dotfiles.git
  dot clone --overwrite -b laptop git@github.com:user/dotfiles.git`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cloneAndInstall(args, root, dot)
	},
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneOverwrite, "overwrite", false, "backup conflicting files and replace them")
	rootCmd.AddCommand(cloneCmd)
}

func cloneAndInstall(args []string, root, dot string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dot clone [git-clone-flags] <repo-url>")
	}

	entries, err := os.ReadDir(dot)
	if err == nil && len(entries) > 0 {
		return fmt.Errorf("%s already exists and is not empty; remove it first or use 'dot install'", dot)
	}

	gitArgs := append([]string{"clone"}, args...)
	gitArgs = append(gitArgs, dot)

	fmt.Printf("Cloning into %s...\n", dot)
	gitClone := exec.Command("git", gitArgs...)
	gitClone.Stdout = os.Stdout
	gitClone.Stderr = os.Stderr
	if err := gitClone.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Println("Installing symlinks...")
	return install(root, dot, overwrite)
}
