package dbs

import (
	"time"
)

type User struct {
	ID   uint64 `storm:"increment"`
	Name string `storm:"index,unique"`
}

type Token struct {
	ID      string
	User    uint64    `storm:"index"`
	Created time.Time `storm:"index"`
}

type Pin struct {
	ID     uint64 `storm:"increment"`
	UserID uint64 `storm:"index"`
	Name   string `storm:"index"`
	CID    string `storm:"index"`
}
