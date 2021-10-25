package db

import (
	"context"
	"fmt"

	"github.com/DeedleFake/sips/ent"
	_ "github.com/lib/pq"
)

// OpenAndMigrate opens the database and performs an auto-migration on it.
func OpenAndMigrate(ctx context.Context, driver, source string, opts ...ent.Option) (*ent.Client, error) {
	entc, err := ent.Open(driver, source, opts...)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	err = entc.Schema.Create(ctx)
	if err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	return entc, nil
}
