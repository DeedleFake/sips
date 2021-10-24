package cmd

import (
	"fmt"
	"regexp"
	"time"

	"github.com/DeedleFake/sips/internal/dbs"
	"github.com/spf13/cobra"
)

var validUserRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var usersCmd = &cobra.Command{
	Use:   "users <subcommand>",
	Short: "administrate users",
	Long: `Adds, removes, and does other administrative tasks for users in the
database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	addCmd := &cobra.Command{
		Use:   "add <username>",
		Short: "add a new user",
		Long:  `Adds a new user.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			if !validUserRE.MatchString(args[0]) {
				return fmt.Errorf("invalid username: %q", args[0])
			}

			tx, err := db.Begin(true)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			user := dbs.User{
				Created: time.Now(),
				Name:    args[0],
			}
			err = tx.Save(&user)
			if err != nil {
				return fmt.Errorf("save user: %w", err)
			}

			fmt.Printf("Successfully created user %q.\nNew user ID: %v\n", user.Name, user.ID)

			return tx.Commit()
		},
	}

	// TODO: Add a command for removing users, possibly along with their
	// pins.

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all existing users",
		Long:  `Lists all registered users in the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			var users []dbs.User
			err = db.All(&users)
			if err != nil {
				return fmt.Errorf("get users: %w", err)
			}

			for _, user := range users {
				fmt.Printf("%v: %q\n", user.ID, user.Name)
			}

			return nil
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm <names...>",
		Short: "remove users from the database",
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
				// TODO: Add the ability to also delete tokens and/or pins for
				// each user.

				var user dbs.User
				err = tx.One("Name", arg, &user)
				if err != nil {
					return fmt.Errorf("find user %q: %w", arg, err)
				}

				err = tx.DeleteStruct(&user)
				if err != nil {
					return fmt.Errorf("delete user %q: %w", arg, err)
				}
			}

			return tx.Commit()
		},
	}

	usersCmd.AddCommand(
		addCmd,
		listCmd,
		rmCmd,
	)
}
