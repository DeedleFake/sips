package cmd

import (
	"fmt"
	"time"

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
	var addFlags struct {
		User string
		Name string
	}
	addCmd := &cobra.Command{
		Use:   "add --user <username> --name <name> <CID>",
		Short: "add a pin to just the database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			tx, err := db.Begin(true)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			var user dbs.User
			err = tx.One("Name", addFlags.User, &user)
			if err != nil {
				return fmt.Errorf("find user: %w", err)
			}

			pin := dbs.Pin{
				Created: time.Now(),
				User:    user.ID,
				Name:    addFlags.Name,
				CID:     args[0],
			}
			err = tx.Save(&pin)
			if err != nil {
				return fmt.Errorf("save pin: %w", err)
			}

			fmt.Printf("New pin ID: %x\n", pin.ID)

			return tx.Commit()
		},
	}
	addCmd.Flags().StringVar(&addFlags.User, "user", "", "pin owner")
	addCmd.MarkFlagRequired("user")
	addCmd.Flags().StringVar(&addFlags.Name, "name", "", "name to identify pin with in the database")
	addCmd.MarkFlagRequired("name")

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

	var rmFlags struct {
		Force bool
	}
	rmCmd := &cobra.Command{
		Use:   "rm <names...>",
		Short: "remove pins from the database",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			tx, err := db.Begin(true)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			for _, arg := range args {
				var pins []dbs.Pin
				err = db.Find("Name", arg, &pins)
				if err != nil {
					return fmt.Errorf("get pins: %w", err)
				}
				if (len(pins) > 1) && !rmFlags.Force {
					return fmt.Errorf("found %v pins but no --force flag given", len(pins))
				}

				for _, pin := range pins {
					err = tx.DeleteStruct(&pin)
					if err != nil {
						return fmt.Errorf("delete pin %v: %w", pin.ID, err)
					}
				}
			}

			return tx.Commit()
		},
	}
	rmCmd.Flags().BoolVar(&rmFlags.Force, "force", false, "allow deletion of multiple matching pins per name")

	pinsCmd.AddCommand(
		addCmd,
		listCmd,
		rmCmd,
	)
}
