package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/modrinth-cli/internal/manifest"
	"github.com/silask7188/modrinth-cli/internal/modrinth"
	"github.com/spf13/cobra"
)

var dest string // may be empty; “auto” when omitted

var addCmd = &cobra.Command{
	Use:   "add <slug|url>",
	Short: "Add a Modrinth project to the manifest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := filepath.Join(gameDir, manifestRel)
		m, err := manifest.Load(path)
		if err != nil {
			return err
		}

		slug := modrinth.ParseSlug(args[0])
		fmt.Printf("Resolving %s …\n", slug)

		if err := m.Add(cmd.Context(), slug, dest); err != nil {
			return err
		}
		if err := m.Save(); err != nil {
			return err
		}
		if dest == "" {
			// manifest.Add chose the folder; look it up to print a nice message
			for _, e := range m.Enabled() {
				if e.Slug == slug {
					dest = e.Dest
				}
			}
		}
		fmt.Printf("Added %s -> %s\n", slug, dest)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(
		&dest,
		"to",
		"", // no default ⇒ auto-infer
		"destination folder (mods, resourcepacks, shaderpacks). Leave blank to infer from project type.",
	)
}
