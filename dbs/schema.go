package dbs

import (
	"time"

	"github.com/DeedleFake/sips"
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

type Job struct {
	ID      uint64 `storm:"increment"`
	Created time.Time
	Pin     uint64 `storm:"index,unique"`
	Mode    JobMode
	Data    sips.Pin
}

type JobMode int

const (
	ModeAdd JobMode = iota
	ModeUpdate
	ModeDelete
)
