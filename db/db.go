package db

import (
	"context"
	"fmt"

	"github.com/DeedleFake/sips/ent"
	_ "github.com/lib/pq"
)

// OpenAndMigrate opens the database and performs an auto-migration on
// it.
func OpenAndMigrate(ctx context.Context, driver, source string, opts ...ent.Option) (*ent.Client, error) {
	entc, err := Open(driver, source, opts...)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	err = entc.Schema.Create(ctx)
	if err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	return entc, nil
}

// Open opens the database. It is exactly equivalent to calling
// ent.Open(), but it is preferred so that this package, and its
// dependencies, are always imported.
func Open(driver, source string, opts ...ent.Option) (*ent.Client, error) {
	return ent.Open(driver, source, opts...)
}
