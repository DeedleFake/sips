package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sipsctl",
	Short: "sipsctl is an admin utility for SIPS",
	Long:  `An admin utility for SIPS, the Simple IPFS Pinning Service.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
