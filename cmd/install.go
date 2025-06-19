package cmd

import (
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/installer"
	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download / update everything that is enabled",
	RunE: func(cmd *cobra.Command, _ []string) error {
		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			println("Manifest not found. Create one with 'mod init --mc [version] --loader [loader]\n")
			return err
		}
		inst, err := installer.New(gameDir, m)
		if err != nil {
			return err
		}
		if err := inst.Install(cmd.Context()); err != nil {
			return err
		}
		// fmt.Println("All done!")
		return nil
	},
}
