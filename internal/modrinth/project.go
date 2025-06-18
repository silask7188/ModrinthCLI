package modrinth

import (
	"encoding/json"
)

// Project represents one element from the /search endpoint.
type Project struct {
	Author               string   `json:"author"`                // "jellysquid3"
	Categories           []string `json:"categories"`            // ["fabric", "neoforge", ...]
	ClientSide           string   `json:"client_side"`           // "required" | "optional" | "unsupported"
	Color                int      `json:"color"`                 // 8703084
	DateCreated          string   `json:"date_created"`          // "2021-01-03T00:53:34.185936Z"
	DateModified         string   `json:"date_modified"`         // "2025-04-04T00:11:16.442076Z"
	Description          string   `json:"description"`           // short description
	DisplayCategories    []string `json:"display_categories"`    // Same as Categories in most cases
	Downloads            int      `json:"downloads"`             // 55 905 812
	FeaturedGallery      *string  `json:"featured_gallery"`      // may be nil
	Follows              int      `json:"follows"`               // 24 330
	IconURL              string   `json:"icon_url"`              // 96×96 icon
	LatestVersion        string   `json:"latest_version"`        // "oZOSEhyy"
	License              License  `json:"license"`               // "LicenseRef-Polyform-Shield-1.0.0"
	ProjectID            string   `json:"project_id"`            // "AANobbMI"
	ProjectType          string   `json:"project_type"`          // "mod", "modpack", etc.
	ServerSide           string   `json:"server_side"`           // "required" | "optional" | "unsupported"
	Slug                 string   `json:"slug"`                  // "sodium"
	Title                string   `json:"title"`                 // "Sodium"
	Versions             []string `json:"versions"`              // ["1.21.5", "1.21.4", ...]
	Body                 string   `json:"body"`                  // body text
	Status               string   `json:"status"`                // status of Project
	RequestedStatus      string   `json:"requested_status"`      // requested status of project?
	AdditionalCategories []string `json:"additional_categories"` // searchable but nonprimary
	IssuesUrl            string   `json:"issues_url"`            // github/issues
	SourceUrl            string   `json:"source_url"`            // github
	WikiUrl              string   `json:"wiki_url"`              // github/wiki
	DiscordUrl           string   `json:"discord_url"`           // discord
	Id                   string   `json:"id"`                    // base62 (?) string
	Team                 string   `json:"team"`                  // team id (base62?)
	// i cant be bothered for the rest
}

type ProjectQuery struct {
	Id   string
	Slug string
}

type License struct {
	Id   string  `json:"id"`   // LGPL-3.0-or-later
	Name string  `json:"name"` // Gnu Lesser .... long name
	Url  *string `json:"url"`  // url to license, nullable
}

func (l *License) UnmarshalJSON(data []byte) error {
	// case 1: just a string
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, &l.Id)
	}

	// case 2: full license
	type alias License
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*l = License(tmp)
	return nil
}

