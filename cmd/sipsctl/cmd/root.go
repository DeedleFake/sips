package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sipsctl",
	Short: "sipsctl is an admin utility for SIPS",
	Long:  `An admin utility for SIPS, the Simple IPFS Pinning Service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
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
}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
