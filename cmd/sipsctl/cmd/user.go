package cmd

import "github.com/spf13/cobra"

var userCmd = &cobra.Command{
	Use:   "user <subcommand>",
	Short: "administrate users",
	Long: `Adds, removes, and does other administrative tasks for users in the
database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
