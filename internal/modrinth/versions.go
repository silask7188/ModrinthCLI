// internal/modrinth/versions.go
package modrinth

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
)

// ────────────────────────────────────────────────────────────────
// Types the installer expects
// ────────────────────────────────────────────────────────────────

// Hashes holds file checksums.
type Hashes struct {
	SHA1 string `json:"sha1"`
}

// File is one downloadable binary / resource-pack / shader-pack.
type File struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Primary  bool   `json:"primary"`
	Hashes   Hashes `json:"hashes"`
}

// Version represents an element of /project/{slug}/version
type Version struct {
	ID            string `json:"id"`
	VersionNumber string `json:"version_number"`
	DatePublished string `json:"date_published"`
	Files         []File `json:"files"`
}

// ────────────────────────────────────────────────────────────────
// Live API calls (no stubs).  Works with your existing doJSON helper.
// ────────────────────────────────────────────────────────────────

// @brief ProjectVersions fetches all versions for a given project slug.
// @param ctx context for cancellation
// @param slug project slug (e.g. "sodium")
// @param gameVersion optional Minecraft version filter (e.g. "1.18.1")
// @param loader optional loader filter (e.g. "fabric")
// @return list of versions or error
func (c *Client) ProjectVersions(
	ctx context.Context,
	slug string,
	gameVersion string, // e.g. "1.18.1"  – empty = no filter
	loader string, // e.g. "fabric"  – empty = no filter
) ([]Version, error) {

	params := url.Values{}
	if gameVersion != "" {
		params.Add("game_versions", fmt.Sprintf(`["%s"]`, gameVersion))
	}
	if loader != "" {
		params.Add("loaders", fmt.Sprintf(`["%s"]`, loader))
	}

	path := fmt.Sprintf("project/%s/version", url.PathEscape(slug))
	out, err := getJSON[[]Version](ctx, c, path, params)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// @brief Version fetches a single version by its ID.
// @param ctx context for cancellation
// @param id version ID (e.g. "SKIBIDI")
// @return Version instance or error
func (c *Client) Version(ctx context.Context, id string) (*Version, error) {
	path := fmt.Sprintf("version/%s", url.PathEscape(id))
	return getJSON[Version](ctx, c, path, nil) // already *Version
}

// @brief ParseSlug extracts the project slug from a Modrinth URL.
// @param s URL or slug string (e.g. "https://modrinth.com/mod/sodium")
// @return slug string (e.g. "sodium") or the original string if no slug found
func ParseSlug(s string) string {
	// Fast-path: already a slug (no slash)
	if !regexp.MustCompile(`/`).MatchString(s) {
		return s
	}
	re := regexp.MustCompile(`modrinth\.com/(?:mod|plugin|datapack|shaderpack|resourcepack)/([^/?#]+)`)
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1]
	}
	return s // fallback – let the API 404 if it’s invalid
}
