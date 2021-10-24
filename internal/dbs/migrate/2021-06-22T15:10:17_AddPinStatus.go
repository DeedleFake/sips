package migrate

import (
	"fmt"
	"time"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/internal/dbs"
	"github.com/asdine/storm"
)

func init() {
	v, err := time.Parse(VersionLayout, "2021-06-22T15:10:17")
	if err != nil {
		panic(fmt.Errorf("parse generated migration version %q: %w", "2021-06-22T15:10:17", err))
	}

	register(v, func(db storm.Node) error {
		var pins []dbs.Pin
		err := db.All(&pins)
		if err != nil {
			return fmt.Errorf("get pins: %w", err)
		}

		for _, pin := range pins {
			if pin.Status != "" {
				continue
			}

			pin.Status = sips.Queued
			err = db.Update(&pin)
			if err != nil {
				return fmt.Errorf("update pin %v: %w", pin.ID, err)
			}
		}

		return nil
	})
}
