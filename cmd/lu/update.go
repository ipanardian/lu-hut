package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/internal/updater"
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	var force bool

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update lu to the latest version",
		Long: `Check for the latest version of lu-hut and update if a newer version is available.

This command will:
  1. Check GitHub releases for the latest version
  2. Download the appropriate binary for your system
  3. Replace the current binary with the new version
  4. Verify the installation

The current binary will be backed up during the update process.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			color.Cyan("Checking for updates...")

			release, err := updater.GetLatestVersion()
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			currentVersion := updater.GetCurrentVersion()
			latestVersion := release.TagName

			fmt.Printf("Current version: %s\n", color.YellowString(currentVersion))
			fmt.Printf("Latest version:  %s\n", color.CyanString(latestVersion))

			if !updater.IsNewerVersion(currentVersion, latestVersion) {
				if force {
					color.Yellow("\nForcing reinstall of %s...", latestVersion)
				} else {
					color.Green("\n✓ You are already running the latest version!")
					return nil
				}
			} else {
				color.Green("\n→ New version available!")
			}

			fmt.Println()

			if err := updater.PerformUpdate(release); err != nil {
				return fmt.Errorf("update failed: %w", err)
			}

			fmt.Println()
			color.Green("Update completed successfully!")
			color.Cyan("Please restart your terminal or run 'hash -r' to use the new version.")

			return nil
		},
	}

	updateCmd.Flags().BoolVarP(&force, "force", "f", false, "force reinstall even if already on latest version")

	updateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println()
		color.Cyan("lu update - Update lu to the latest version")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  lu update [flags]")
		fmt.Println()
		fmt.Println("FLAGS:")
		fmt.Println("  -f, --force    force reinstall even if already on latest version")
		fmt.Println("      --help     help for update")
		fmt.Println()
		fmt.Println("DESCRIPTION:")
		fmt.Println("  This command will:")
		fmt.Println("    1. Check GitHub releases for the latest version")
		fmt.Println("    2. Download the appropriate binary for your system")
		fmt.Println("    3. Replace the current binary with the new version")
		fmt.Println("    4. Verify the installation")
		fmt.Println()
		fmt.Println("  The current binary will be backed up during the update process.")
		fmt.Println()
	})

	return updateCmd
}

func newVersionCommand() *cobra.Command {
	var checkUpdate bool

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the current version of lu-hut and optionally check for updates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentVersion := updater.GetCurrentVersion()

			fmt.Printf("lu-hut version %s\n", color.CyanString(currentVersion))
			fmt.Printf("OS/Arch: %s\n", color.YellowString(updater.GetBinaryName()))

			if checkUpdate {
				fmt.Println()
				color.Cyan("Checking for updates...")

				release, err := updater.GetLatestVersion()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
					return nil
				}

				latestVersion := release.TagName
				fmt.Printf("Latest version:  %s\n", color.CyanString(latestVersion))

				if updater.IsNewerVersion(currentVersion, latestVersion) {
					fmt.Println()
					color.Yellow("→ New version available!")
					color.Cyan("Run 'lu update' to upgrade")
				} else {
					fmt.Println()
					color.Green("✓ You are running the latest version!")
				}
			}

			return nil
		},
	}

	versionCmd.Flags().BoolVarP(&checkUpdate, "check", "c", false, "check for available updates")

	versionCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println()
		color.Cyan("lu version - Show version information")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  lu version [flags]")
		fmt.Println()
		fmt.Println("FLAGS:")
		fmt.Println("  -c, --check    check for available updates")
		fmt.Println("      --help     help for version")
		fmt.Println()
		fmt.Println("EXAMPLES:")
		fmt.Println("  lu version")
		fmt.Println("  lu version --check")
		fmt.Println()
	})

	return versionCmd
}
