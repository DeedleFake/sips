package dbs

import (
	"fmt"

	"github.com/asdine/storm"
)

func Open(path string) (*storm.DB, error) {
	db, err := storm.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	init := func(t interface{}) {
		if err != nil {
			return
		}
		err = db.Init(t)
	}
	init(new(User))
	init(new(Token))
	init(new(Pin))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return db, nil
}
