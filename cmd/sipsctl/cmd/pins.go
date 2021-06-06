package cmd

import (
	"fmt"

	"github.com/DeedleFake/sips/dbs"
	"github.com/spf13/cobra"
)

var pinsCmd = &cobra.Command{
	Use:   "pins <subcommand>",
	Short: "administrate pins",
	Long:  `Adds, removes, and lists pins without needing to use the HTTP API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all pins in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			var pins []dbs.Pin
			err = db.All(&pins)
			if err != nil {
				return fmt.Errorf("get pins: %w", err)
			}

			for _, pin := range pins {
				fmt.Printf("%v: %v as %q\n", pin.ID, pin.CID, pin.Name)
			}

			return nil
		},
	}

	var syncFlags struct {
		Delete bool
	}
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "synchronize pins to IPFS",
		Long:  `Synchronizes pin status to actual IPFS pins.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			panic("Not implemented.")
		},
	}
	syncCmd.Flags().BoolFlag(&syncFlags.Delete, "delete", false, "delete missing pins from IPFS")

	pinsCmd.AddCommand(
		listCmd,
		syncCmd,
	)
}
