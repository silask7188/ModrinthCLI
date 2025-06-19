package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/silask7188/ModrinthCLI/internal/modrinth"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for mods, resource packs, or shaders on Modrinth",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no search query provided")
		}
		query := args[0]
		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}
		facets := modrinth.Facets{
			MinecraftVersion: m.Minecraft.Version,
			Loader:           m.Minecraft.Loader,
		}
		params := modrinth.SearchParams{
			Query:  query,
			Facets: facets,
			Offset: 0,
			Limit:  10,
		}
		c, err := modrinth.New("https://api.modrinth.com/v2/")
		if err != nil {
			return fmt.Errorf("failed to create Modrinth client: %w", err)
		}

		result, err := c.Search(cmd.Context(), params)
		if err != nil {
			return fmt.Errorf("failed to search Modrinth: %w", err)
		}
		results := result.Hits
		if len(results) == 0 {
			return fmt.Errorf("no results found for '%s'", query)
		}
		fmt.Printf("Found %d results for '%s':\n", len(results), query)
		for _, result := range results {
			fmt.Printf("- %-30s (%-1s)\n", result.Slug, result.Title)
		}
		return nil
	},
}
