package cmd

import (
	"fmt"
	"regexp"

	"github.com/DeedleFake/sips/db"
	"github.com/DeedleFake/sips/ent/user"
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

			if !validUserRE.MatchString(args[0]) {
				return fmt.Errorf("invalid username: %q", args[0])
			}

			u, err := tx.User.Create().
				SetName(args[0]).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create user: %w", err)
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			fmt.Printf("Added user %q\n", u.Name)
			fmt.Printf("  ID: %d\n", u.ID)

			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all existing users",
		Long:  `Lists all registered users in the database.`,
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

			users, err := tx.User.Query().All(ctx)
			if err != nil {
				return fmt.Errorf("query users: %w", err)
			}

			for _, u := range users {
				fmt.Printf("%v: %q\n", u.ID, u.Name)
			}

			return nil
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm <names...>",
		Short: "remove users from the database",
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

			n, err := tx.User.Delete().
				Where(user.NameIn(args...)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("delete users: %w", err)
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			fmt.Printf("Deleted %d users\n", n)

			return nil
		},
	}

	usersCmd.AddCommand(
		addCmd,
		listCmd,
		rmCmd,
	)
}
