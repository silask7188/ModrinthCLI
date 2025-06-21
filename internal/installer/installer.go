package installer

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/silask7188/ModrinthCLI/internal/modrinth"
)

/*
--------------------------------------------------
  STRUCT & CONSTRUCTOR
--------------------------------------------------
*/

// Installer resolves, downloads, and places jars / packs.
type Installer struct {
	gameDir string
	man     *manifest.Manifest
	api     *modrinth.Client
	http    *http.Client
	concur  int // worker count
}

// @brief New creates a new Installer instance.
// @param gameDir path to the game directory
// @param man Manifest instance
// @return Installer instance or error
func New(gameDir string, man *manifest.Manifest) (*Installer, error) {
	api, err := modrinth.New("https://api.modrinth.com/v2/")
	if err != nil {
		return nil, err
	}
	return &Installer{
		gameDir: gameDir,
		man:     man,
		api:     api,
		http: &http.Client{
			Timeout: 45 * time.Second,
		},
		concur: 4, // default – can expose flag later
	}, nil
}

/*
--------------------------------------------------
  PUBLIC ENTRY-POINTS
--------------------------------------------------
*/

// @brief Install downloads and installs all enabled mods.
// @param ctx context for cancellation
// @return error if any
func (ins *Installer) Install(ctx context.Context) error {
	sections := []*[]manifest.Entry{
		&ins.man.Mods,
		&ins.man.ResourcePacks,
		&ins.man.Shaders,
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.SetLimit(ins.concur)

	for _, section := range sections {
		for i := range *section {
			ent := &(*section)[i]
			if !ent.Enable {
				continue
			}
			grp.Go(func(ent *manifest.Entry) func() error {
				return func() error {
					return ins.installOne(ctx, ent)
				}
			}(ent))
		}
	}
	return grp.Wait()
}

// Update record – used by PlanUpdates().
type Update struct {
	Entry          manifest.Entry
	CurrentVersion string
	TargetVersion  string
}

// @brief PlanUpdates checks for updates to enabled mods.
// @param ctx context for cancellation
// @return slice of Update records or error
func (ins *Installer) PlanUpdates(ctx context.Context) ([]Update, error) {
	var out []Update
	for i, e := range ins.man.Enabled() {
		latest, err := ins.resolveVersion(ctx, e)
		if err != nil {
			return nil, err
		}
		cur := e.Version
		if cur != latest {
			out = append(out, Update{
				Entry:          ins.man.Enabled()[i],
				CurrentVersion: cur,
				TargetVersion:  latest,
			})
		}
	}
	return out, nil
}

/*
--------------------------------------------------
  INTERNAL HELPERS
--------------------------------------------------
*/

// @brief installOne resolves the version, downloads the file, and places it in the game directory.
// @param ctx context for cancellation
// @param e manifest entry to install
// @return error if any
func (ins *Installer) installOne(ctx context.Context, e *manifest.Entry) error {
	verID, err := ins.resolveVersion(ctx, *e)
	if err != nil {
		return err
	}

	file, sha1sum, err := ins.fileForVersion(ctx, verID)
	if err != nil {
		return err
	}

	destDir := filepath.Join(ins.gameDir, e.Dest)
	if err = os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, file.Filename)

	// check if the expected file is already present and valid
	if ok, _ := fileExistsWithSHA1(destPath, sha1sum); ok {
		return nil
	}

	// fallback: search for any file with the correct checksum
	if found, err := findFileByChecksum(destDir, sha1sum); err == nil {
		// found correct file under a different name — update manifest
		e.Filename = found
		e.Checksum = sha1sum
		e.Version = verID
		_ = ins.man.Save() // silently save fix
		return nil
	}

	// file not found download
	tmp, err := ins.download(ctx, file.URL, sha1sum)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err = backupIfExists(destPath); err != nil {
		return err
	}

	if err = os.Rename(tmp, destPath); err != nil {
		return fmt.Errorf("failed to move file to final destination: %w", err)
	}

	// record updated manifest data
	e.Checksum = sha1sum
	e.Filename = file.Filename
	e.Version = verID
	if err := ins.man.Save(); err != nil {
		return err
	}

	fmt.Printf("[+] %s -> %s (%s)\n", e.Slug, e.Dest, e.VersionNumber)
	return nil
}

// @brief resolveVersion fetches the latest compatible version ID for a given entry.
// @param ctx context for cancellation
// @param e manifest entry to resolve
// @return version ID or error
func (ins *Installer) resolveVersion(ctx context.Context, e manifest.Entry) (string, error) {
	var vers []modrinth.Version
	var err error
	switch e.Dest {
	case "mods":
		vers, err = ins.api.ProjectVersions(
			ctx,
			e.Slug,
			ins.man.Minecraft.Version,
			ins.man.Minecraft.Loader,
		)
	case "resourcepacks":
		vers, err = ins.api.ProjectVersions(
			ctx,
			e.Slug,
			ins.man.Minecraft.Version,
			"",
		)
	case "shaderpacks":
		vers, err = ins.api.ProjectVersions(
			ctx,
			e.Slug,
			ins.man.Minecraft.Version,
			"",
		)
	default:
		return "", fmt.Errorf("unknown type %q for %s", e.Dest, e.Slug)
	}
	if err != nil {
		return "", err
	}
	if len(vers) == 0 {
		return "", fmt.Errorf("no compatible versions for %s", e.Slug)
	}
	// ensure newest first by published date
	sort.Slice(vers, func(i, j int) bool {
		return vers[i].DatePublished > vers[j].DatePublished
	})
	return vers[0].ID, nil
}

// @brief fileForVersion fetches the primary file for a given version ID.
// @param ctx context for cancellation
// @param verID version ID to fetch
// @return File instance and its SHA1 hash, or error
func (ins *Installer) fileForVersion(ctx context.Context, verID string) (*modrinth.File, string, error) {
	v, err := ins.api.Version(ctx, verID)
	if err != nil {
		return nil, "", err
	}
	for _, f := range v.Files {
		if f.Primary {
			return &f, f.Hashes.SHA1, nil
		}
	}
	f := v.Files[0]
	return &f, f.Hashes.SHA1, nil
}

/*
--------------------------------------------------
  LOW-LEVEL UTILS
--------------------------------------------------
*/

// @brief download fetches a file from the given URL and verifies its SHA256 hash.
// @param ctx context for cancellation
// @param url URL to download from
// @param wantSHA expected SHA256 hash of the file
// @return path to the downloaded file or error
func (ins *Installer) download(ctx context.Context, url, wantSHA string) (string, error) {

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := ins.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	tmp, err := os.CreateTemp("", "mr-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	h := sha1.New()
	if _, err = io.Copy(io.MultiWriter(tmp, h), res.Body); err != nil {
		return "", err
	}
	var got string
	if got = hex.EncodeToString(h.Sum(nil)); got != wantSHA {
		return "", fmt.Errorf("sha1 mismatch for %s (want %s, got %s)", url, wantSHA, got)
	}
	// fmt.Println("Downloading:", url)
	// fmt.Println("Expecting SHA-1:", wantSHA)
	// fmt.Println("Actual SHA-1:", got)
	return tmp.Name(), nil
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

func backupIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		ts := time.Now().Format("20060102-150405")
		return os.Rename(path, path+"."+ts+".bak")
	}
	return nil
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
