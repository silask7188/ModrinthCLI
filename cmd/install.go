package cmd

import (
	"path/filepath"

	"github.com/silask7188/modrinth-cli/internal/installer"
	"github.com/silask7188/modrinth-cli/internal/manifest"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download / update everything that is enabled",
	RunE: func(cmd *cobra.Command, _ []string) error {
		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return err
		}
		inst, err := installer.New(gameDir, m)
		if err := inst.Install(cmd.Context()); err != nil {
			return err
		}
		// fmt.Println("All done!")
		return nil
	},
}
