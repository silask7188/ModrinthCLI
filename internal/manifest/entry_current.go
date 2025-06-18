package manifest

import (
	"os"
	"path/filepath"
	"strings"
)

// @brief Returns the current file for this entry in the given game directory.
// @param gameDir The base directory where the game files are located.
// @return The filename of the current file, or an empty string if not found.
func (e Entry) Current(gameDir string) string {
	dir := filepath.Join(gameDir, e.Dest)
	ents, _ := os.ReadDir(dir)
	prefix := e.Slug + "-"
	for _, fi := range ents {
		if fi.Type().IsRegular() && strings.HasPrefix(fi.Name(), prefix) {
			return fi.Name()
		}
	}
	return ""
}
