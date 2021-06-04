package dbs

import (
	"time"

	"github.com/asdine/storm"
)

type User struct {
	ID   uint64 `storm:"increment"`
	Name string `storm:"unique"`
	Pins storm.Node
}

type Token struct {
	ID      string
	User    uint64
	Created time.Time
}

type Pin struct {
	ID   uint64 `storm:"increment"`
	Name string
	CID  string
}
