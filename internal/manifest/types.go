package manifest

type Entry struct {
	Slug          string `json:"slug"`
	Version       string `json:"version"`
	VersionNumber string `json:"version_number"` // human-readable
	Dest          string `json:"dest"`
	Filename      string `json:"filename"`
	Enable        bool   `json:"enable"`
}

type Minecraft struct {
	Loader        string `json:"loader"`            // forge,fabric,neoforge
	LoaderVersion string `json:"loader_version"`    // fabric "1.6.5"
	Version       string `json:"minecraft_version"` // minecraft 1.21.6
}

type Manifest struct {
	Schema        int       `json:"schema"` // modrinth-cli ver
	Minecraft     Minecraft `json:"minecraft"`
	Mods          []Entry   `json:"mods"`
	ResourcePacks []Entry   `json:"resource_packs"`
	Shaders       []Entry   `json:"shaders"`
	path          string    // absolute
}
