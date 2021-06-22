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

func Run(db *storm.DB) error {
	var current time.Time
	err := db.Get(Bucket, VersionKey, &current)
	if (err != nil) && !errors.Is(err, storm.ErrNotFound) {
		return fmt.Errorf("get current schema version: %w", err)
	}

	var start int
	if err != nil {
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
