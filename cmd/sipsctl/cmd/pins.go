package cmd

import (
	"fmt"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/db"
	"github.com/DeedleFake/sips/ent/pin"
	"github.com/DeedleFake/sips/ent/user"
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
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, "postgres", rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer entc.Close()

			tx, err := entc.Tx(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			u, err := tx.User.Query().
				Where(user.Name(addFlags.User)).
				Only(ctx)
			if err != nil {
				return fmt.Errorf("find user: %w", err)
			}

			pin, err := tx.Pin.Create().
				SetUser(u).
				SetName(addFlags.Name).
				SetCID(args[0]).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create pin: %w", err)
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			fmt.Printf("New pin ID: %v\n", pin.ID)

			return nil
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
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, "postgres", rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer entc.Close()

			tx, err := entc.Tx(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			pins, err := tx.Pin.Query().All(ctx)
			if err != nil {
				return fmt.Errorf("query pins: %w", err)
			}

			for _, pin := range pins {
				fmt.Printf("%v: %v as %q (%v)\n", pin.ID, pin.CID, pin.Name, pin.Status)
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
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, "postgres", rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer entc.Close()

			tx, err := entc.Tx(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			for _, name := range args {
				_, err := tx.Pin.Delete().
					Where(pin.Name(name)).
					Exec(ctx)
				if err != nil {
					return fmt.Errorf("delete pin %q: %w", name, err)
				}
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			return nil
		},
	}
	rmCmd.Flags().BoolVar(&rmFlags.Force, "force", false, "allow deletion of multiple matching pins per name")

	var setstatusFlags struct {
		Status string
	}
	setstatusCmd := &cobra.Command{
		Use:   "setstatus <pin IDs...>",
		Short: "manually sets the status of pins",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, "postgres", rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer entc.Close()

			tx, err := entc.Tx(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			for _, id := range args {
				err := tx.Pin.Update().
					Where(pin.ID(id)).
					SetStatus(sips.RequestStatus(setstatusFlags.Status)).
					Exec(ctx)
				if err != nil {
					return fmt.Errorf("update pin %v: %w", id, err)
				}
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			return nil
		},
	}
	setstatusCmd.Flags().StringVar(&setstatusFlags.Status, "status", string(sips.Queued), "status to reset pins to")

	pinsCmd.AddCommand(
		addCmd,
		listCmd,
		rmCmd,
		setstatusCmd,
	)
}
