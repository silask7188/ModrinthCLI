package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/silask7188/ModrinthCLI/internal/modrinth"
	"github.com/spf13/cobra"
)

var (
	includeMods          bool
	includeResourcePacks bool
	includeShaders       bool
	limit                int
	offset               int = 1
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for mods, resource packs, or shaders on Modrinth. By default, searches for all",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no search query provided")
		}
		query := args[0]

		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}

		var facets modrinth.Facets
		facets.MinecraftVersion = m.Minecraft.Version

		// Append all specified project types
		var types []string
		if includeMods {
			types = append(types, "mod")
		}
		if includeResourcePacks {
			types = append(types, "resourcepack")
		}
		if includeShaders {
			types = append(types, "shader")
		}
		if len(types) > 0 {
			facets.ProjectType = types
		}

		params := modrinth.SearchParams{
			Query:  query,
			Facets: facets,
			Offset: offset - 1,
			Limit:  limit,
		}

		client, err := modrinth.New("https://api.modrinth.com/v2/")
		if err != nil {
			return fmt.Errorf("failed to create Modrinth client: %w", err)
		}

		result, err := client.Search(cmd.Context(), params)
		if err != nil {
			return fmt.Errorf("failed to search Modrinth: %w", err)
		}

		if len(result.Hits) == 0 {
			return fmt.Errorf("no results found for '%s'", query)
		}

		// Group results
		var mods, packs, shaders []modrinth.Project
		for _, hit := range result.Hits {
			switch hit.ProjectType {
			case "mod":
				if contains(hit.Categories, m.Minecraft.Loader) && contains(hit.Versions, m.Minecraft.Version) {
					mods = append(mods, hit)
				}
			case "resourcepack":
				packs = append(packs, hit)
			case "shader":
				shaders = append(shaders, hit)
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if len(mods) > 0 {
			fmt.Fprintln(w, "MODS")
			for _, p := range mods {
				fmt.Fprintf(w, "%s\t%s\thttps://modrinth.com/mod/%s\n", p.Slug, p.Title, p.Slug)
			}
		}
		if len(packs) > 0 {
			fmt.Fprintln(w, "\nRESOURCE PACKS")
			for _, p := range packs {
				fmt.Fprintf(w, "%s\t%s\thttps://modrinth.com/resourcepack/%s\n", p.Slug, p.Title, p.Slug)
			}
		}
		if len(shaders) > 0 {
			fmt.Fprintln(w, "\nSHADERS")
			for _, p := range shaders {
				fmt.Fprintf(w, "%s\t%s\thttps://modrinth.com/shader/%s\n", p.Slug, p.Title, p.Slug)
			}
		}
		w.Flush()
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVarP(&includeMods, "mod", "m", false, "Include mods in the search")
	searchCmd.Flags().BoolVarP(&includeResourcePacks, "resourcepack", "r", false, "Include resource packs in the search")
	searchCmd.Flags().BoolVarP(&includeShaders, "shaders", "s", false, "Include shaders in the search")
	searchCmd.Flags().IntVarP(&limit, "limit", "l", 30, "Number of results to return (default 30)")
	// searchCmd.Flags().IntVarP(&offset, "page", "p", 1, "Page (default 1)")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
