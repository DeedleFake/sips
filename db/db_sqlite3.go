//go:build sqlite3
// +build sqlite3

package db

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	drivers = append(drivers, "sqlite3")
}
