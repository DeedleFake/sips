package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/DeedleFake/sips/db"
	"github.com/DeedleFake/sips/internal/cli"
	"github.com/spf13/cobra"
)

// errEarlyExit is a signal value to indicate that a command exited
// early cleanly, despite returning an error. Cobra is weird.
var errEarlyExit = errors.New("early exit")

var rootCmd = &cobra.Command{
	Use:           "sipsctl",
	Short:         "sipsctl is an admin utility for SIPS",
	Long:          `An admin utility for SIPS, the Simple IPFS Pinning Service.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if rootFlags.DBDriver == "list" {
			fmt.Println("Available database drivers:")
			for _, t := range db.Drivers() {
				fmt.Printf("  %s\n", t)
			}
			return errEarlyExit
		}

		dbpath, _, err := cli.ExpandConfig(rootFlags.DBPath)
		if err != nil {
			return fmt.Errorf("expand database path: %w", err)
		}
		rootFlags.DBPath = dbpath
		return nil
	},
}

var rootFlags struct {
	DBDriver string
	DBPath   string
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&rootFlags.DBDriver,
		"dbdriver",
		"postgres",
		"database driver to use (\"list\" to show available)",
	)
	rootCmd.PersistentFlags().StringVar(
		&rootFlags.DBPath,
		"db",
		"host=/var/run/postgresql dbname=sips",
		"database connection string ($CONFIG will be replaced with user config dir path)",
	)

	rootCmd.AddCommand(
		tokensCmd,
		usersCmd,
		pinsCmd,
		migrateCmd,
	)
}

func ExecuteContext(ctx context.Context) error {
	err := rootCmd.ExecuteContext(ctx)
	if errors.Is(err, errEarlyExit) {
		return nil
	}
	return err
}
