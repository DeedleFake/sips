package db

import (
	"context"
	"fmt"

	"github.com/DeedleFake/sips/ent"
	dbs "github.com/DeedleFake/sips/internal/bolt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

// MigrateFromBolt migrates data from the old BoltDB system to the new ent-based one.
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
		var pins []dbs.Pin
		err := bolt.Select(q.Eq("UserID", user.ID)).Find(&pins)
		if err != nil {
			return fmt.Errorf("get BoltDB pins for user %d: %w", user.ID, err)
		}
		epins := make([]*ent.Pin, len(pins))
		for i, pin := range pins {
			epin, err := tx.Pin.Create().
				SetName(pin.Name).
				SetStatus(pin.Status).
				SetOrigins(pin.Origins).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create ent pin: %w", err)
			}
			epins[i] = epin
		}

		var tokens []dbs.Token
		err = bolt.Select(q.Eq("UserID", user.ID)).Find(&tokens)
		if err != nil {
			return fmt.Errorf("get BoltDB tokens for user %d: %w", user.ID, err)
		}
		etokens := make([]*ent.Token, len(pins))
		for i, token := range tokens {
			etoken, err := tx.Token.Create().
				SetToken(token.ID).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create ent token: %w", err)
			}
			etokens[i] = etoken
		}

		_, err = tx.User.Create().
			SetName(user.Name).
			AddTokens(etokens...).
			AddPins(epins...).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create ent user for BoltDB user %d: %w", user.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
