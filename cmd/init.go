package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	mcVersion     string
	loader        string
	loaderVersion string
	forge         bool
	fabric        bool
	neoforge      bool
	quilt         bool
)

var initCmd = &cobra.Command{
	Use:   "init [DIRECTORY]",
	Short: "Create  a new project with empty manifest",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := gameDir
		if len(args) == 1 {
			dir = args[0]
		} else if len(args) == 0 {
			dir = "."
		}
		var mc manifest.Minecraft
		mc.Loader = loader
		mc.LoaderVersion = loaderVersion
		mc.Version = mcVersion

		count := 0
		if forge {
			count++
		}
		if fabric {
			count++
		}
		if neoforge {
			count++
		}
		if quilt {
			count++
		}
		if count > 1 {
			return fmt.Errorf("multiple loaders specified, use only one of --forge, --fabric, --neoforge, --quilt")
		}
		if count == 1 && loader != "" {
			fmt.Println("Loader already specified, ignoring flags.")
		}

		if mc.Loader == "" {
			if forge {
				mc.Loader = "forge"
			} else if fabric {
				mc.Loader = "fabric"
			} else if neoforge {
				mc.Loader = "neoforge"
			} else if quilt {
				mc.Loader = "quilt"
			} else {
				return fmt.Errorf("loader is required, use --loader [loader] or one of the flags --forge, --fabric, --neoforge, --quilt")
			}
		}

		if mc.Version == "latest" {
			mc.Version = "1.21.6"
		}
		if mc.Version == "" {
			return fmt.Errorf("minecraft version is required, use --mc [version/latest]")
		}
		if mc.Loader == "" {
			return fmt.Errorf("loader is required, use --loader [loader]")
		}
		if mc.LoaderVersion == "" {
			fmt.Println("Loader version not specified, using latest")
			mc.LoaderVersion = "latest"
		}

		path := filepath.Join(dir, manifestRel)
		m := manifest.New(path, mc)

		if err := m.Save(); err != nil {
			return err
		}
		fmt.Printf("Created %s\n", path)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&mcVersion, "mc", "", "Minecraft version")
	initCmd.Flags().StringVar(&loader, "loader", "", "loader (vanilla, fabric, quilt, neoforge...)")
	initCmd.Flags().StringVar(&loaderVersion, "loader-version", "", "loader version (optional, latest by default)")
	initCmd.Flags().BoolVar(&forge, "forge", false, "use Forge loader")
	initCmd.Flags().BoolVar(&fabric, "fabric", false, "use Fabric loader")
	initCmd.Flags().BoolVar(&neoforge, "neoforge", false, "use NeoForge loader")
	initCmd.Flags().BoolVar(&quilt, "quilt", false, "use Quilt loader")
}
