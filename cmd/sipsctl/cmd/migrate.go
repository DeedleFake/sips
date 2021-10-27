package cmd

import (
	"fmt"

	"github.com/DeedleFake/sips/db"
	dbs "github.com/DeedleFake/sips/internal/bolt"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate <subcommand>",
	Short: "modify database migration state",
	Long:  `Run migrations or perform similar tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	var fromboltArgs struct {
		BoltDBPath string
	}
	fromboltCmd := &cobra.Command{
		Use:   "frombolt",
		Short: "migrate from old BoltDB database",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			log.Infof("opening BoltDB database")
			bolt, err := dbs.Open(fromboltArgs.BoltDBPath)
			if err != nil {
				return fmt.Errorf("open BoltDB database: %w", err)
			}
			defer bolt.Close()

			log.Infof("opening ent database")
			entc, err := db.OpenAndMigrate(ctx, rootFlags.DBDriver, rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open PostgreSQL database: %w", err)
			}
			defer entc.Close()

			log.Infof("migrating from BoltDB database to ent database")
			err = db.MigrateFromBolt(ctx, entc, bolt)
			if err != nil {
				return fmt.Errorf("migrate: %w", err)
			}

			log.Infof("migration complete")
			return nil
		},
	}
	fromboltCmd.Flags().StringVar(&fromboltArgs.BoltDBPath, "boltdb", "", "path to old BoltDB database")
	fromboltCmd.MarkFlagRequired("boltdb")

	migrateCmd.AddCommand(
		fromboltCmd,
	)
}
