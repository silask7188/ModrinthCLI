package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable <slug1> <slug2> ...",
	Short: "Disable mods in the manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no slugs provided")
		}

		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return err
		}

		var disabled []string
		for _, slug := range args {
			if err := m.Disable(gameDir, slug); err != nil {
				return fmt.Errorf("failed to disable %s: %w", slug, err)
			}
			disabled = append(disabled, slug)
		}

		if err := m.Save(); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}

		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "DISABLED\tSLUG")
		for _, slug := range disabled {
			fmt.Fprintf(tw, "✓\t%s\n", strings.TrimSpace(slug))
		}
		return tw.Flush()
	},
}
