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
		dbpath, _, err := cli.ExpandConfig(globalFlags.DBPath)
		if err != nil {
			return fmt.Errorf("expand database path: %w", err)
		}
		globalFlags.DBPath = dbpath
		return nil
	},
}

var globalFlags struct {
	DBPath string
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&globalFlags.DBPath,
		"db",
		"$CONFIG/sips/database.db",
		"path to database ($CONFIG will be replaced with user config dir path)",
	)

	rootCmd.AddCommand(tokenCmd)
	rootCmd.AddCommand(userCmd)
}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
