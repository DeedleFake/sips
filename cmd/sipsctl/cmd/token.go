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

func init() {
	var addUserFlag string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "generate a new auth token",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(globalFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			var user dbs.User
			err = db.One("Name", addUserFlag, &user)
			if err != nil {
				return fmt.Errorf("get user %q: %v", addUserFlag, err)
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
	addCmd.Flags().StringVar(&addUserFlag, "user", "", "user to generate a token for")
	addCmd.MarkFlagRequired("user")

	tokenCmd.AddCommand(
		addCmd,
	)
}
