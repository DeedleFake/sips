package dbs

import (
	"time"
)

type User struct {
	ID      uint64 `storm:"increment"`
	Created time.Time
	Name    string `storm:"index,unique"`
}

type Token struct {
	ID      string
	Created time.Time
	User    uint64 `storm:"index"`
}

type Pin struct {
	ID      uint64 `storm:"increment"`
	Created time.Time
	User    uint64 `storm:"index"`
	Name    string `storm:"index"`
	CID     string `storm:"index"`
}
