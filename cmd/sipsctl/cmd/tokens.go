package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/DeedleFake/sips/db"
	"github.com/DeedleFake/sips/ent/token"
	"github.com/DeedleFake/sips/ent/user"
	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens <subcommand>",
	Short: "administrate auth tokens",
	Long:  `Administrate authenitication tokens in the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var tokenFlags struct {
	User string
}

func init() {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "generate a new auth token",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, rootFlags.DBDriver, rootFlags.DBPath)
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
				Where(user.Name(tokenFlags.User)).
				Only(ctx)
			if err != nil {
				return fmt.Errorf("find user: %w", err)
			}

			var buf [256]byte
			_, err = rand.Read(buf[:])
			if err != nil {
				return fmt.Errorf("generate random bytes for token: %v", err)
			}
			sum := sha256.Sum256(buf[:])

			tok, err := tx.Token.Create().
				SetUser(u).
				SetToken(base64.URLEncoding.EncodeToString(sum[:])).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create token: %w", err)
			}

			fmt.Println(tok.Token)

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			return nil
		},
	}
	addCmd.MarkPersistentFlagRequired("user")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, rootFlags.DBDriver, rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer entc.Close()

			tx, err := entc.Tx(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			q := tx.Token.Query()
			if tokenFlags.User != "" {
				q = q.Where(token.HasUserWith(user.Name(tokenFlags.User)))
			}
			toks, err := q.WithUser().All(ctx)
			if err != nil {
				return fmt.Errorf("list tokens: %w", err)
			}

			for _, tok := range toks {
				userName := "<no user>"
				if tok.Edges.User != nil {
					userName = tok.Edges.User.Name
				}
				fmt.Printf("%v %v\n", tok.Token, userName)
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			return nil
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm <tokens...>",
		Short: "remove a token from the database, thus invalidating it",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			entc, err := db.OpenAndMigrate(ctx, rootFlags.DBDriver, rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer entc.Close()

			tx, err := entc.Tx(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback()

			n, err := tx.Token.Delete().
				Where(token.TokenIn(args...)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("delete tokens: %w", err)
			}

			fmt.Printf("Deleted %v tokens\n", n)

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			return nil
		},
	}

	tokensCmd.AddCommand(
		addCmd,
		listCmd,
		rmCmd,
	)
	tokensCmd.PersistentFlags().StringVar(&tokenFlags.User, "user", "", "user that token is associated with")
}
