package cmd

import "github.com/spf13/cobra"

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "administrate auth tokens",
	Long:  `Administrate authenitication tokens in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	tokenCmd.AddCommand(
		&cobra.Command{
			Use:   "add",
			Short: "generate a new auth token",
			Run: func(cmd *cobra.Command, args []string) {
				panic("Not implemented.")
			},
		},
	)
}
