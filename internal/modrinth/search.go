package modrinth

import (
	"fmt"
	"net/url"
	"strings"
)

type SearchResponse struct {
	Hits      []Project `json:"hits"`       // how many results
	Offset    int       `json:"offset"`     // pagination
	Limit     int       `json:"limit"`      // pagination
	TotalHits int       `json:"total_hits"` // pagination
}

type SearchParams struct {
	Query  string // "sodium"
	Facets Facets // ["categories=fabric", "project_type=mod"]
	Offset int    // pagination
	Limit  int    //pagination
}

type Facets struct {
	ProjectType []string `json:"project_type,omitempty"` // ["mod", "resource_pack", "shader"]
	Loader      string   `json:"loader,omitempty"`       // ["fabric", "forge", "neoforge"]
	// LoaderVersion    string   `json:"loader_version,omitempty"`    // ["1.6.5", "1.7.10"]
	MinecraftVersion string   `json:"minecraft_version,omitempty"` // ["1.21.6", "1.20.5"]
	Category         []string `json:"categories,omitempty"`        // ["tech", "exploration", "quality-of-life"]
}

// @brief Values returns the URL parameters for the search request.
// @return url.Values with the search parameters
func (p SearchParams) Values() url.Values {
	v := make(url.Values)

	if p.Query != "" {
		v.Set("query", p.Query)
	}

	// Build facets array - each facet should be in its own array for AND logic
	var facetArrays []string

	for _, pt := range p.Facets.ProjectType {
		facetArrays = append(facetArrays, fmt.Sprintf(`["project_type:%s"]`, pt))
	}
	if p.Facets.Loader != "" {
		facetArrays = append(facetArrays, fmt.Sprintf(`["categories:%s"]`, p.Facets.Loader))
	}
	// if p.Facets.LoaderVersion != "" {
	// 	facetArrays = append(facetArrays, fmt.Sprintf(`["versions:%s"]`, p.Facets.LoaderVersion))
	// }
	if p.Facets.MinecraftVersion != "" {
		facetArrays = append(facetArrays, fmt.Sprintf(`["versions:%s"]`, p.Facets.MinecraftVersion))
	}

	// Handle categories - if multiple categories, put them in one array for OR logic
	if len(p.Facets.Category) > 0 {
		var categoryFacets []string
		for _, cat := range p.Facets.Category {
			categoryFacets = append(categoryFacets, fmt.Sprintf(`"categories:%s"`, cat))
		}
		facetArrays = append(facetArrays, fmt.Sprintf(`[%s]`, strings.Join(categoryFacets, ",")))
	}

	// Join all facet arrays into the final facets parameter
	if len(facetArrays) > 0 {
		v.Set("facets", fmt.Sprintf("[%s]", strings.Join(facetArrays, ",")))
	}

	if p.Offset > 0 {
		v.Set("offset", fmt.Sprint(p.Offset))
	}
	if p.Limit > 0 {
		v.Set("limit", fmt.Sprint(p.Limit))
	}
	return v
}
