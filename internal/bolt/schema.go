package dbs

import (
	"time"

	"github.com/DeedleFake/sips"
)

// User reprsents a user in the database.
type User struct {
	ID      uint64 `storm:"increment"`
	Created time.Time
	Name    string `storm:"index,unique"`
}

// Token represents an auth token in the database.
type Token struct {
	ID      string
	Created time.Time
	User    uint64 `storm:"index"`
}

// Pin represents a pin entry in the database.
type Pin struct {
	ID      uint64 `storm:"increment"`
	Created time.Time
	User    uint64             `storm:"index"`
	Status  sips.RequestStatus `storm:"index"`
	Name    string             `storm:"index"`
	CID     string             `storm:"index"`
	Origins []string
}
