// Package migrate implements very simple database migrations.
//
// Migrations are stored by timestamp and are run in ascending order.
package migrate

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/asdine/storm"
)

const (
	Bucket     = "schema"
	VersionKey = "version"
)

// A Migration is a function that is run in order to implement changes
// in the database. Each migration is atomic, and all of the changes
// made will be rolled back in the event of an error.
type Migration func(storm.Node) error

var migrations []migrationRegistration

type migrationRegistration struct {
	V time.Time
	M Migration
}

func register(version time.Time, migration Migration) {
	i := sort.Search(len(migrations), func(i int) bool {
		return version.Unix() <= migrations[i].V.Unix()
	})
	migrations = append(
		migrations[:i],
		append(
			[]migrationRegistration{{V: version, M: migration}},
			migrations[i:]...,
		)...,
	)
}

func run(db *storm.DB, migration migrationRegistration) error {
	tx, err := db.Begin(true)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback()

	err = migration.M(tx)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	err = tx.Set(Bucket, VersionKey, migration.V)
	if err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}

	return tx.Commit()
}

// CurrentVersion returns the current schema version stored in the
// database. If no current version information is found, the returned
// error is nil and the returned boolean is false.
func CurrentVersion(db storm.Node) (time.Time, bool, error) {
	var current time.Time
	err := db.Get(Bucket, VersionKey, &current)
	if (err != nil) && !errors.Is(err, storm.ErrNotFound) {
		return time.Time{}, false, fmt.Errorf("get: %w", err)
	}
	return current, err == nil, nil
}

// Run runs migrations against a database.
func Run(db *storm.DB) error {
	current, found, err := CurrentVersion(db)
	if err != nil {
		return fmt.Errorf("current version: %w", err)
	}

	var start int
	if found {
		start = sort.Search(len(migrations), func(i int) bool {
			return current.Unix() <= migrations[i].V.Unix()
		})
		if (start >= len(migrations)) || !migrations[start].V.Equal(current) {
			return fmt.Errorf("current schema version is %v but is not registered", current)
		}
		start++ // Start with the migration after the current one.
	}

	for _, migration := range migrations[start:] {
		err := run(db, migration)
		if err != nil {
			return fmt.Errorf("migration %v: %w", migration.V, err)
		}
	}

	return nil
}
