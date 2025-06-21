package manifest

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/modrinth"
)

const Version = 1

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
	m.baseDir = filepath.Dir(path)
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
		baseDir:   filepath.Dir(path),
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

	switch prj.ProjectType {
	case "mod":
		for i := range m.Mods {
			if m.Mods[i].Slug == slug {
				m.Mods[i].Dest = "mods"
				m.Mods[i].Version = latest.ID
				m.Mods[i].VersionNumber = latest.VersionNumber
				m.Mods[i].Enable = true
				return nil
			}
		}
		// Not found, add new
		m.Mods = append(m.Mods, Entry{
			Slug:          slug,
			Dest:          "mods",
			Version:       latest.ID,
			VersionNumber: latest.VersionNumber,
			Enable:        true,
		})
		return nil
	case "resourcepack":
		for i := range m.ResourcePacks {
			if m.ResourcePacks[i].Slug == slug {
				m.ResourcePacks[i].Dest = "resourcepacks"
				m.ResourcePacks[i].Version = latest.ID
				m.ResourcePacks[i].VersionNumber = latest.VersionNumber
				m.ResourcePacks[i].Enable = true
				return nil
			}
		}
		m.ResourcePacks = append(m.ResourcePacks, Entry{
			Slug:          slug,
			Dest:          "resourcepacks",
			Version:       latest.ID,
			VersionNumber: latest.VersionNumber,
			Enable:        true,
		})
		return nil
	case "shader":
		for i := range m.Shaders {
			if m.Shaders[i].Slug == slug {
				m.Shaders[i].Dest = "shaderpacks"
				m.Shaders[i].Version = latest.ID
				m.Shaders[i].VersionNumber = latest.VersionNumber
				m.Shaders[i].Enable = true
				return nil
			}
		}
		m.Shaders = append(m.Shaders, Entry{
			Slug:          slug,
			Dest:          "shaderpacks",
			Version:       latest.ID,
			VersionNumber: latest.VersionNumber,
			Enable:        true,
		})
		return nil
	default:
		return fmt.Errorf("unknown project type %q for slug %q", prj.ProjectType, slug)
	}
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
func toggleDisabled(gameDir string, entries []Entry, slug string, wantEnable bool) error {
	for i := range entries {
		ent := &entries[i]
		if ent.Slug != slug {
			continue
		}
		ent.Enable = wantEnable

		dir := filepath.Join(gameDir, ent.Dest)
		filename := ent.Filename

		// fallback: scan directory for a file with matching checksum
		if filename == "" {
			found, err := findFileByChecksum(dir, ent.Checksum)
			if err != nil {
				return fmt.Errorf("no filename recorded for %s and no match by checksum: %w", slug, err)
			}
			filename = found
			ent.Filename = filename // optional: heal the manifest
		}

		enabledPath := filepath.Join(dir, filename)
		disabledPath := enabledPath + ".disabled"

		var from, to string
		if wantEnable {
			from, to = disabledPath, enabledPath
		} else {
			from, to = enabledPath, disabledPath
		}

		if _, err := os.Stat(from); os.IsNotExist(err) {
			return fmt.Errorf("file %s not found; expected %s", slug, from)
		}

		if err := os.Rename(from, to); err != nil {
			return fmt.Errorf("failed to toggle %s: %w", slug, err)
		}

		return nil
	}
	return fmt.Errorf("slug %s not in manifest", slug)
}

// @brief remove an item from a specific section of the manifest.
// @param section the section to remove from (mods, resourcepacks, shaders)
// @param slug modrinth project slug
// @return error if the item was not found or could not be removed
func (m *Manifest) RemoveFromSection(section string, slug string) error {
	var mods *[]Entry
	switch section {
	case "mods":
		mods = &m.Mods
	case "resourcepacks":
		mods = &m.ResourcePacks
	case "shaders":
		mods = &m.Shaders
	default:
		return fmt.Errorf("unknown section %q", section)
	}

	for i := range *mods {
		if (*mods)[i].Slug == slug {
			*mods = append((*mods)[:i], (*mods)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("slug %s not found in section %s", slug, section)
}

// @brief Remove an entry from the manifest and delete its file from disk.
// @param gameDir path to the game directory
// @param slug modrinth project slug
// @return error if the entry was not found or could not be removed
func (m *Manifest) Remove(gameDir, slug string) error {
	sections := []struct {
		name    string
		entries *[]Entry
	}{
		{"mods", &m.Mods},
		{"resourcepacks", &m.ResourcePacks},
		{"shaders", &m.Shaders},
	}

	for _, sec := range sections {
		for i := range *sec.entries {
			if (*sec.entries)[i].Slug != slug {
				continue
			}

			entry := (*sec.entries)[i]

			// Resolve filename
			filename := entry.Filename
			if filename == "" {
				found, err := findFileByChecksum(filepath.Join(gameDir, entry.Dest), entry.Checksum)
				if err != nil {
					return fmt.Errorf("could not find file for %s by filename or checksum: %w", slug, err)
				}
				filename = found
				entry.Filename = found
			}

			// Delete file from disk
			path := filepath.Join(gameDir, entry.Dest, filename)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove file %q: %w", path, err)
			}

			// Remove from manifest
			*sec.entries = append((*sec.entries)[:i], (*sec.entries)[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("slug %s not found in any section", slug)
}

func findFileByChecksum(dir, wantHash string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ok, err := fileExistsWithSHA1(path, wantHash)
		if err != nil {
			return "", err
		}
		if ok {
			return entry.Name(), nil
		}
	}
	return "", fmt.Errorf("no file with checksum %s found in %s", wantHash, dir)
}

// @brief computeSHA1 computes the SHA1 hash of a file.
func computeSHA1(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// @brief fileExistsWithSHA1 checks if a file exists and matches the given SHA1 hash.
// @param path file path to check
// @param want expected SHA1 hash
// @return true if file exists and matches, false if not, or error if any
func fileExistsWithSHA1(path, want string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	h := sha1.New()
	if _, err = io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == want, nil
}

// @brief check for filename updates by comparing sha1 hashes.
// @return error if any filename does not match its checksum
// @return a list of filenames that do not match their checksums
func (m *Manifest) CheckFilenames() ([]string, error) {
	var mismatches []string
	var changed bool

	check := func(entries *[]Entry) {
		for i := range *entries {
			ent := &(*entries)[i]

			if ent.Filename == "" {
				mismatches = append(mismatches, fmt.Sprintf("%s: no filename recorded, may not exist yet\n", ent.Slug))
				continue
			}

			dir := filepath.Join(m.baseDir, ent.Dest)
			expected := filepath.Join(dir, ent.Filename)

			// check if the expected file matches the checksum ! yipee
			ok, err := fileExistsWithSHA1(expected, ent.Checksum)
			if err != nil {
				mismatches = append(mismatches, fmt.Sprintf("%s: %s\n", ent.Slug, err))
				continue
			}
			if ok {
				continue // all good
			}

			// try to find a file in the folder with the right checksum
			foundName, err := findFileByChecksum(dir, ent.Checksum)
			if err == nil {
				mismatches = append(mismatches,
					fmt.Sprintf("%s: filename changed from %s to %s\n", ent.Slug, ent.Filename, foundName))
				ent.Filename = foundName
				changed = true
				continue
			}

			// if the original file exists, but has a different checksum !! uh oh!! unless not a mod
			if _, err := os.Stat(expected); err == nil {
				actual, err := computeSHA1(expected)
				if err != nil {
					mismatches = append(mismatches, fmt.Sprintf("%s: failed to hash %s: %s\n", ent.Slug, expected, err))
				} else {
					mismatches = append(mismatches,
						fmt.Sprintf("%s: %s checksum mismatch (want %s, got %s)\n",
							ent.Slug, ent.Filename, ent.Checksum, actual))
				}
			} else {
				// file not found at all, and not recoverable by hash
				mismatches = append(mismatches,
					fmt.Sprintf("%s: file %s not found and no match by checksum\n", ent.Slug, expected))
			}
		}
	}

	check(&m.Mods)
	check(&m.ResourcePacks)
	check(&m.Shaders)

	if changed {
		err := m.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save manifest after filename updates: %w", err)
		}
	}

	if len(mismatches) > 0 {
		return mismatches, fmt.Errorf("found %d filename or checksum mismatches", len(mismatches))
	}
	return nil, nil
}
