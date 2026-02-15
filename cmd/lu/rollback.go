package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/internal/updater"
	"github.com/spf13/cobra"
)

func newRollbackCommand() *cobra.Command {
	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to the previous version",
		Long: `Restore the previous version of lu-hut from backup.

This command will:
  1. Check if a backup version exists
  2. Restore the backup binary
  3. Verify the rollback was successful

The backup is created automatically during the update process.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			color.Cyan("Checking for backup version...")

			if err := updater.PerformRollback(); err != nil {
				return fmt.Errorf("rollback failed: %w", err)
			}

			fmt.Println()
			color.Green("Rollback completed successfully!")
			color.Cyan("Please restart your terminal or run 'hash -r' to use the restored version.")

			return nil
		},
	}

	rollbackCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println()
		color.Cyan("lu rollback - Rollback to the previous version")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  lu rollback")
		fmt.Println()
		fmt.Println("FLAGS:")
		fmt.Println("      --help     help for rollback")
		fmt.Println()
		fmt.Println("DESCRIPTION:")
		fmt.Println("  This command will:")
		fmt.Println("    1. Check if a backup version exists")
		fmt.Println("    2. Restore the backup binary")
		fmt.Println("    3. Verify the rollback was successful")
		fmt.Println()
		fmt.Println("  The backup is created automatically during the update process.")
		fmt.Println()
	})

	return rollbackCmd
}
