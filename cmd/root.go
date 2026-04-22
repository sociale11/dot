package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	root string // defaults to $HOME
	dot  string // defaults to $HOME/.local/share/dot
)

var rootCmd = &cobra.Command{
	Use:   "dot",
	Short: "A minimal dotfile manager",
	Long:  "dot tracks dotfiles by moving them into a central directory and symlinking them back.",
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
	rootCmd.PersistentFlags().StringVar(&dot, "dot", filepath.Join(home, ".local/share/dot"), "dot storage directory")
}

// use rel path
