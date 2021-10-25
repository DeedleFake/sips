package cmd

import (
	"context"
	"fmt"

	"github.com/DeedleFake/sips/internal/cli"
	"github.com/spf13/cobra"
)

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
		dbpath, _, err := cli.ExpandConfig(rootFlags.DBPath)
		if err != nil {
			return fmt.Errorf("expand database path: %w", err)
		}
		rootFlags.DBPath = dbpath
		return nil
	},
}

var rootFlags struct {
	DBPath string
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&rootFlags.DBPath,
		"db",
		"host=/var/run/postgres dbname=sips sslmode=disable",
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
	return rootCmd.ExecuteContext(ctx)
}
