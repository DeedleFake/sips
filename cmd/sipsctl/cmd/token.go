package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/DeedleFake/sips/dbs"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
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
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			var user dbs.User
			err = db.One("Name", tokenFlags.User, &user)
			if err != nil {
				return fmt.Errorf("get user %q: %v", tokenFlags.User, err)
			}

			var buf [256]byte
			_, err = rand.Read(buf[:])
			if err != nil {
				return fmt.Errorf("generate random bytes for token: %v", err)
			}
			sum := sha256.Sum256(buf[:])

			tok := dbs.Token{
				ID:      base64.URLEncoding.EncodeToString(sum[:]),
				User:    user.ID,
				Created: time.Now(),
			}
			err = db.Save(&tok)
			if err != nil {
				return fmt.Errorf("save token to database: %v", err)
			}

			fmt.Println(tok.ID)

			return nil
		},
	}
	addCmd.MarkPersistentFlagRequired("user")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			// TODO: Only show tokens for a given user if the user flag was
			// provided.

			var tokens []dbs.Token
			err = db.All(&tokens)
			if err != nil {
				return fmt.Errorf("get tokens: %w", err)
			}

			userCache := make(map[uint64]string)
			for _, token := range tokens {
				user, ok := userCache[token.User]
				if !ok {
					var u dbs.User
					err := db.One("ID", token.User, &u)
					if err != nil {
						return fmt.Errorf("get user %v: %w", token.User, err)
					}

					userCache[token.User] = u.Name
					user = u.Name
				}

				fmt.Printf("%v %v\n", token.ID, user)
			}

			return nil
		},
	}

	tokenCmd.AddCommand(
		addCmd,
		listCmd,
	)
	tokenCmd.PersistentFlags().StringVar(&tokenFlags.User, "user", "", "user that token is associated with")
}
