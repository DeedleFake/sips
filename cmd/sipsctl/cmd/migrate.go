package cmd

import (
	"fmt"
	"time"

	"github.com/DeedleFake/sips/internal/dbs"
	"github.com/DeedleFake/sips/internal/dbs/migrate"
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
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			err = migrate.Run(db)
			if err != nil {
				return fmt.Errorf("run migrations: %w", err)
			}

			return nil
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <version>",
		Short: "manually set the current database schema version",
		Long: fmt.Sprintf(`Sets the current database schema version. The version provided must
match the layout %q. As a special case, it may instead
be the word "clear", in which case the version info is removed from
the database completely.`, migrate.VersionLayout),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			if args[0] == "clear" {
				err = db.Delete(migrate.Bucket, migrate.VersionKey)
				if err != nil {
					return fmt.Errorf("delete version: %w", err)
				}
				return nil
			}

			version, err := time.Parse(migrate.VersionLayout, args[0])
			if err != nil {
				return fmt.Errorf("parse version: %w", err)
			}

			err = db.Set(migrate.Bucket, migrate.VersionKey, version)
			if err != nil {
				return fmt.Errorf("set version: %w", err)
			}

			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "get the current database schema version",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := dbs.Open(rootFlags.DBPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			current, found, err := migrate.CurrentVersion(db)
			if err != nil {
				return fmt.Errorf("get current version: %w", err)
			}

			if !found {
				fmt.Println("no version info in database")
				return nil
			}

			fmt.Println(current.Format(migrate.VersionLayout))
			return nil
		},
	}

	migrateCmd.AddCommand(
		runCmd,
		setCmd,
		getCmd,
	)
}
