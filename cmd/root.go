package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	gameDir     string
	manifestRel string
	rootCmd     = &cobra.Command{
		Use:   "mod",
		Short: "Minecraft Mod/Resourcepack/Shader Manager",
	}
)

func Execute() {
	// global flags
	rootCmd.PersistentFlags().StringVar(&gameDir, "dir", ".", "path to project directory")
	rootCmd.PersistentFlags().StringVar(&manifestRel, "manifest", "project.json", "manifest filename")

	// subcommands
	rootCmd.AddCommand(initCmd, addCmd, listCmd, installCmd, updateCmd, enableCmd, disableCmd, removeCmd, searchCmd, checkCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
