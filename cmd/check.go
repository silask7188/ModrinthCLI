package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/installer"
	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for issues in the manifest",
	RunE: func(cmd *cobra.Command, _ []string) error {
		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}

		inst, err := installer.New(gameDir, m)
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}
		if inst == nil {
			return fmt.Errorf("installer is nil, cannot check manifest")
		}

		res, err := m.CheckFilenames()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "something wrong! : %v\n", err)
		}
		if len(res) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Manifest is valid ✓")
		} else {
			for _, r := range res {
				fmt.Fprint(cmd.OutOrStdout(), r)
			}
			print("Please fix the issues above and try again.\nIf the file was renamed, you are fine.\n")
		}
		return nil
	},
}

func init() {
	checkCmd.SilenceUsage = true
}
