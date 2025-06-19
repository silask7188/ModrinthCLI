package cmd

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
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
		// println("MODS")
		fmt.Fprintln(tw, "\tITEM\tVERSION\tTYPE")
		for _, e := range m.Mods {
			en := "✓"
			if !e.Enable {
				en = "x"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", en, e.Slug, e.VersionNumber, e.Dest)
		}
		// println("RESOURCE PACKS")
		for _, e := range m.ResourcePacks {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "", e.Slug, e.VersionNumber, e.Dest)
		}
		// println("SHADERS")
		for _, e := range m.Shaders {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "", e.Slug, e.VersionNumber, e.Dest)
		}
		return tw.Flush()
	},
}
