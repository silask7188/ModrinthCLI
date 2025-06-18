package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable <slug1> <slug2> ...",
	Short: "Enable mods in the manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no slugs provided")
		}

		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return err
		}

		var enabled []string
		for _, slug := range args {
			if err := m.Enable(gameDir, slug); err != nil {
				return fmt.Errorf("failed to enable %s: %w", slug, err)
			}
			enabled = append(enabled, slug)
		}

		if err := m.Save(); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}

		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "ENABLED\tSLUG")
		for _, slug := range enabled {
			fmt.Fprintf(tw, "✓\t%s\n", strings.TrimSpace(slug))
		}
		return tw.Flush()
	},
}
