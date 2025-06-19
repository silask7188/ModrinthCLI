package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/silask7188/ModrinthCLI/internal/modrinth"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <slug1> <slug2> ...",
	Short: "Disable items in the manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no slugs provided")
		}

		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return err
		}

		var removed []string
		for _, slug := range args {
			if err := m.Remove(gameDir, modrinth.ParseSlug(slug)); err != nil {
				return fmt.Errorf("failed to remove %s: %w", slug, err)
			}
			removed = append(removed, slug)
		}

		if err := m.Save(); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}

		fmt.Printf("Removed: %s\n", removed)
		return nil
	},
}
