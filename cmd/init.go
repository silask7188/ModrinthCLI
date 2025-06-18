package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	mcVersion     string
	loader        string
	loaderVersion string
)

var initCmd = &cobra.Command{
	Use:   "init [DIRECTORY]",
	Short: "Create  a new project with empty manifest",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := gameDir
		if len(args) == 1 {
			dir = args[0]
		} else if len(args) == 0 {
			dir = "."
		}
		var mc manifest.Minecraft
		mc.Loader = loader
		mc.LoaderVersion = loaderVersion
		mc.Version = mcVersion
		path := filepath.Join(dir, manifestRel)
		m := manifest.New(path, mc)
		if err := m.Save(); err != nil {
			return err
		}
		fmt.Printf("Created %s\n", path)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&mcVersion, "mc", "1.21.6", "Minecraft version")
	initCmd.Flags().StringVar(&loader, "loader", "neoforge", "loader (fabric, quilt, neoforge)")
}
