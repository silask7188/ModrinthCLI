package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/silask7188/modrinth-cli/internal/modrinth"
)

const Version = 0

// @brief load a manifest from disk
// @param path path to the manifest file
// @return Manifest instance or error
func Load(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m.path = path
	return &m, nil
}

// @brief new manifest instance
// @param path path to the manifest file
// @param mc Minecraft instance with version and loader
// @return Manifest instance
func New(path string, mc Minecraft) *Manifest {
	return &Manifest{
		Schema:    Version,
		Minecraft: mc,
		path:      path,
	}
}

// @brief save the manifest to disk
// @return error if saving failed
func (m *Manifest) Save() error {
	if m.path == "" {
		return errors.New("Manifest path unset")
	}
	b, _ := json.MarshalIndent(m, "", " ")
	return os.WriteFile(m.path, b, 0o644)
}

// @brief add a new entry to the manifest
// @param ctx context for API calls
// @param slug modrinth project slug
// @param dest destination folder (mods, resourcepacks, shaders)
// @return error if the project was not found or could not be added
func (m *Manifest) Add(ctx context.Context, slug, dest string) error {
	cli, err := modrinth.New("https://api.modrinth.com/v2/")
	if err != nil {
		return err
	}

	prj, err := cli.GetProject(ctx, modrinth.ProjectQuery{Slug: slug})
	if err != nil {
		return fmt.Errorf("modrinth project %q not found: %w", slug, err)
	}

	if dest == "" {
		switch prj.ProjectType {
		case "mod":
			dest = "mods"
		case "resourcepack":
			dest = "resourcepacks"
		case "shader":
			dest = "shaderpacks"
		default:
			return fmt.Errorf("cannot infer destination for project type %q; use --to flag", prj.ProjectType)
		}
	}

	gameVer := m.Minecraft.Version
	loader := m.Minecraft.Loader

	needLooseSearch := prj.ProjectType == "resourcepack" || prj.ProjectType == "shader"

	vers, err := cli.ProjectVersions(ctx, slug, gameVer, func() string {
		if needLooseSearch && dest != "mods" {
			return "" // ignore loader for packs/shaders
		}
		return loader
	}())
	if err != nil {
		return err
	}

	if len(vers) == 0 && needLooseSearch {
		vers, err = cli.ProjectVersions(ctx, slug, "", "")
		if err != nil {
			return err
		}
	}

	if len(vers) == 0 {
		return fmt.Errorf("no compatible versions for %s (MC=%s loader=%s)",
			slug, gameVer, loader)
	}
	latest := vers[0] // newest -> oldest

	if prj.ProjectType == "mod" {
		for i := range m.Mods {
			if m.Mods[i].Slug == slug {
				m.Mods[i].Dest = "mods"
				m.Mods[i].Version = latest.ID
				m.Mods[i].VersionNumber = latest.VersionNumber
				m.Mods[i].Enable = true
				return nil
			}
		}
	} else if prj.ProjectType == "resourcepack" {
		for i := range m.ResourcePacks {
			if m.ResourcePacks[i].Slug == slug {
				m.ResourcePacks[i].Dest = "resourcepacks"
				m.ResourcePacks[i].Version = latest.ID
				m.ResourcePacks[i].VersionNumber = latest.VersionNumber
				m.ResourcePacks[i].Enable = true
				return nil
			}
		}
	} else if prj.ProjectType == "shader" {
		for i := range m.Shaders {
			if m.Shaders[i].Slug == slug {
				m.Shaders[i].Dest = "shaderpacks"
				m.Shaders[i].Version = latest.ID
				m.Shaders[i].VersionNumber = latest.VersionNumber
				m.Shaders[i].Enable = true
				return nil
			}
		}
	} else {
		return fmt.Errorf("unknown project type %q for slug %q", prj.ProjectType, slug)
	}

	m.Mods = append(m.Mods, Entry{
		Slug:          slug,
		Version:       latest.ID,
		VersionNumber: latest.VersionNumber,
		Dest:          dest,
		Enable:        true,
	})
	return nil
}

// @brief get all enabled entries in the manifest
// @return slice of enabled entries
func (m *Manifest) Enabled() []Entry {
	out := make([]Entry, 0, len(m.Mods))
	for _, e := range m.Mods {
		if e.Enable {
			out = append(out, e)
		}
	}
	return out
}

// -------------------------------------------------------------------
// Enable / Disable – rename file or folder on disk
// -------------------------------------------------------------------

// @brief Enable turns a mod on (removes .disabled suffix so game sees it).
// @param gameDir path to the game directory
// @param slug modrinth project slug
// @return error if the mod was not found or could not be enabled
func (m *Manifest) Enable(gameDir, slug string) error {
	return toggleDisabled(gameDir, m.Mods, slug, true)
}

// @brief Disable turns a mod off (adds .disabled suffix so game ignores it).
// @param gameDir path to the game directory
// @param slug modrinth project slug
// @return error if the mod was not found or could not be disabled
func (m *Manifest) Disable(gameDir, slug string) error {
	return toggleDisabled(gameDir, m.Mods, slug, false)
}

// @brief toggleDisabled enables or disables a mod by renaming its file or folder.
// @param gameDir path to the game directory
// @param mods slice of manifest entries
// @param slug modrinth project slug
// @param wantEnable true to enable, false to disable
// @return error if the mod was not found or could not be toggled
func toggleDisabled(gameDir string, mods []Entry, slug string, wantEnable bool) error {
	for i := range mods {
		ent := &mods[i]
		if ent.Slug != slug {
			continue
		}
		ent.Enable = wantEnable

		if ent.Filename == "" {
			return fmt.Errorf("no filename recorded for %s; reinstall first", slug)
		}

		dir := filepath.Join(gameDir, ent.Dest)
		oldPath := filepath.Join(dir, ent.Filename)
		disPath := oldPath + ".disabled"

		var from, to string
		if wantEnable {
			from, to = disPath, oldPath
		} else {
			from, to = oldPath, disPath
		}

		if _, err := os.Stat(from); os.IsNotExist(err) {
			return fmt.Errorf("file %s not found; expected %s", slug, from)
		}
		if err := os.Rename(from, to); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("slug %s not in manifest", slug)
}
