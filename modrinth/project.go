package modrinth

// Project represents one element from the /search endpoint.
type Project struct {
	Author            string   `json:"author"`             // "jellysquid3"
	Categories        []string `json:"categories"`         // ["fabric", "neoforge", ...]
	ClientSide        string   `json:"client_side"`        // "required" | "optional" | "unsupported"
	Color             int      `json:"color"`              // 8703084
	DateCreated       string   `json:"date_created"`       // "2021-01-03T00:53:34.185936Z"
	DateModified      string   `json:"date_modified"`      // "2025-04-04T00:11:16.442076Z"
	Description       string   `json:"description"`        // Long mod blurb
	DisplayCategories []string `json:"display_categories"` // Same as Categories in most cases
	Downloads         int      `json:"downloads"`          // 55 905 812
	FeaturedGallery   *string  `json:"featured_gallery"`   // may be nil
	Follows           int      `json:"follows"`            // 24 330
	Gallery           []string `json:"gallery"`            // Slice of image URLs
	IconURL           string   `json:"icon_url"`           // 96×96 icon
	LatestVersion     string   `json:"latest_version"`     // "oZOSEhyy"
	License           string   `json:"license"`            // "LicenseRef-Polyform-Shield-1.0.0"
	ProjectID         string   `json:"project_id"`         // "AANobbMI"
	ProjectType       string   `json:"project_type"`       // "mod", "modpack", etc.
	ServerSide        string   `json:"server_side"`        // "required" | "optional" | "unsupported"
	Slug              string   `json:"slug"`               // "sodium"
	Title             string   `json:"title"`              // "Sodium"
	Versions          []string `json:"versions"`           // ["1.21.5", "1.21.4", ...]
}

type SearchResponse struct {
	Hits              []Project `json:"hits"`
	Offset			  int       `json:"offset"`
	Limit             int       `json:"limit"`
	Totalhits         int       `json:"total_hits"`
}
