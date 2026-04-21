package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	root  string // defaults to $HOME
	dotly string // defaults to $HOME/.local/share/dotly
)

var rootCmd = &cobra.Command{
	Use:   "dotly",
	Short: "A minimal dotfile manager",
	Long:  "dotly tracks dotfiles by moving them into a central directory and symlinking them back.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not determine home directory:", err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&root, "root", home, "root directory (treated as $HOME)")
	rootCmd.PersistentFlags().StringVar(&dotly, "dotly", filepath.Join(home, ".local/share/dotly"), "dotly storage directory")
}
