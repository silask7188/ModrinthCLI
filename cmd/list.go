package cmd

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/silask7188/modrinth-cli/internal/manifest"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show manifest entries",
	RunE: func(cmd *cobra.Command, _ []string) error {
		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return err
		}
		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "\tTYPE\tSLUG")
		for _, e := range m.Mods {
			en := "✓"
			if !e.Enable {
				en = "x"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", en, e.Dest, e.Slug)
		}
		return tw.Flush()
	},
}
