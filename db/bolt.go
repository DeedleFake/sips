package db

import (
	"context"
	"fmt"

	"github.com/DeedleFake/sips/ent"
	dbs "github.com/DeedleFake/sips/internal/bolt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

// MigrateFromBolt migrates data from the old BoltDB system to the new one.
func MigrateFromBolt(ctx context.Context, entc *ent.Client, bolt *storm.DB) error {
	tx, err := entc.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var users []dbs.User
	err = bolt.All(&users)
	if err != nil {
		return fmt.Errorf("get BoltDB users: %w", err)
	}

	for _, user := range users {
		u, err := tx.User.Create().
			SetCreateTime(user.Created).
			SetName(user.Name).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create ent user for BoltDB user %d: %w", user.ID, err)
		}

		var pins []dbs.Pin
		err = bolt.Select(q.Eq("User", user.ID)).Find(&pins)
		if err != nil {
			return fmt.Errorf("get BoltDB pins for user %d: %w", user.ID, err)
		}
		for _, pin := range pins {
			_, err := tx.Pin.Create().
				SetCreateTime(pin.Created).
				SetUser(u).
				SetName(pin.Name).
				SetStatus(pin.Status).
				SetOrigins(pin.Origins).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create ent pin: %w", err)
			}
		}

		var tokens []dbs.Token
		err = bolt.Select(q.Eq("User", user.ID)).Find(&tokens)
		if err != nil {
			return fmt.Errorf("get BoltDB tokens for user %d: %w", user.ID, err)
		}
		for _, token := range tokens {
			_, err := tx.Token.Create().
				SetCreateTime(token.Created).
				SetToken(token.ID).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create ent token: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
