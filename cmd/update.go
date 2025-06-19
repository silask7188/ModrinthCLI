package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/silask7188/ModrinthCLI/internal/installer"
	"github.com/silask7188/ModrinthCLI/internal/manifest"

	"github.com/spf13/cobra"
)

var dryRun bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install newer compatible versions",
	RunE: func(cmd *cobra.Command, _ []string) error {
		m, err := manifest.Load(filepath.Join(gameDir, manifestRel))
		if err != nil {
			return err
		}

		print("Checking for mods not yet installed...\n")
		inst, err := installer.New(gameDir, m)
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}
		if err := inst.Install(cmd.Context()); err != nil {
			return err
		}

		plan, err := inst.PlanUpdates(cmd.Context())
		if err != nil {
			return err
		}
		if len(plan) == 0 {
			fmt.Println("Everything is up-to-date ✓")
			return nil
		}
		print("Checking for updates...\n")
		var total int
		for _, p := range plan {
			if p.CurrentVersion == "" {
				fmt.Printf("[ ] %-20s  %s -> %s (new)\n", p.Entry.Slug, p.CurrentVersion, p.TargetVersion)
				total++
			} else if p.CurrentVersion == p.TargetVersion {
				fmt.Printf("[=] %-20s  %s -> %s (already up-to-date)\n", p.Entry.Slug, p.CurrentVersion, p.TargetVersion)
			} else if p.TargetVersion == "" {
				fmt.Printf("[x] %-20s  %s -> %s (no compatible version found)\n", p.Entry.Slug, p.CurrentVersion, p.TargetVersion)
			} else if p.TargetVersion == "latest" {
				fmt.Printf("[ ] %-20s  %s -> %s (latest)\n", p.Entry.Slug, p.CurrentVersion, p.TargetVersion)
				total++
			} else {
				fmt.Printf("[ ] %-20s  %s -> %s\n", p.Entry.Slug, p.CurrentVersion, p.TargetVersion)
				total++
			}
		}
		fmt.Printf("Found %d updates\n", total)
		if dryRun {
			return nil
		}
		return inst.Install(cmd.Context())
	},
}

func init() {
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show updates without installing")
}
