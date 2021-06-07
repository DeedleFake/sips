package cmd

import (
	"fmt"
	"regexp"
	"time"

	"github.com/DeedleFake/sips/dbs"
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

			user := dbs.User{
				Created: time.Now(),
				Name:    args[0],
			}
			err = db.Save(&user)
			if err != nil {
				return fmt.Errorf("save user: %w", err)
			}

			fmt.Printf("Successfully created user %q.\nNew user ID: %v\n", user.Name, user.ID)

			return nil
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

	usersCmd.AddCommand(
		addCmd,
		listCmd,
	)
}
