// Package dbutil provides utilities for dealing with the database.
//
// The database is stored in a nested key/value pair scheme. The general structure used by this package is as follows:
//
//    - USERS
//      - <user ID>
//        - NAME -> <username>
//        - PINS
//          - <pin name> -> <CID>
//    - TOKENS
//      - <token ID>
//        - USER -> <user ID>
//        - CREATED -> <creation time>
//    - PINS
//      - <pin ID>
//        - NAME -> <name for the pin>
//        - CID -> <CID of pinned data>
package dbutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	userBucket  = "USERS"
	userNameKey = "NAME"
	userPinsKey = "PINS"

	tokenBucket     = "TOKENS"
	tokenUserKey    = "USER"
	tokenCreatedKey = "CREATED"

	pinBucket  = "PINS"
	pinNameKey = "NAME"
	pinCIDKey  = "CID"
)

func Open(dbpath string, createDir bool) (*bbolt.DB, error) {
	if createDir {
		err := os.MkdirAll(filepath.Dir(dbpath), 0770)
		if err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := bbolt.Open(dbpath, 0660, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		createBucket := func(name []byte) {
			if err != nil {
				return
			}
			_, err = tx.CreateBucketIfNotExists(name)
		}

		createBucket([]byte(userBucket))
		createBucket([]byte(tokenBucket))

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("create buckets: %w", err)
	}

	return db, nil
}

type MatchFunc func(key []byte, data *bbolt.Bucket) (bool, error)

func ByKey(key []byte) MatchFunc {
	return func(k []byte, b *bbolt.Bucket) (bool, error) {
		return bytes.Equal(k, key), nil
	}
}

type extractFunc func(key []byte, data *bbolt.Bucket) error

func getData(parent *bbolt.Bucket, match MatchFunc, extract extractFunc) error {
	cursor := parent.Cursor()
	for {
		k, v := cursor.Next()
		if v != nil {
			continue
		}
		if k == nil {
			return ErrNotFound
		}

		ok, err := match(k, parent.Bucket(k))
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		return extract(k, parent.Bucket(k))
	}
}

type User struct {
	ID   uint64
	Name string
	Pins []uint64
}

func GetUser(db *bbolt.DB, match MatchFunc) (user User, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		return getData(tx.Bucket([]byte(userBucket)), match, func(k []byte, b *bbolt.Bucket) error {
			panic("Not implemented.")
		})
	})
	return user, err
}

func UserByID(id uint64) MatchFunc {
	var buf [8]byte
	n := binary.PutUvarint(buf[:], id)
	return ByKey(buf[:n])
}

type Token struct {
	ID     string
	UserID uint64
}

func CreateToken(db *bbolt.DB, userID uint64) (token Token, err error) {
	_, err = GetUser(db, UserByID(userID))
	if err != nil {
		return token, fmt.Errorf("get user %v: %w", userID, err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		panic("Not implemented.")
	})
	return token, err
}

func GetToken(db *bbolt.DB, match MatchFunc) (token Token, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		return getData(tx.Bucket([]byte(tokenBucket)), match, func(k []byte, b *bbolt.Bucket) error {
			rawUserID := b.Get([]byte(tokenUserKey))
			userID, n := binary.Uvarint(rawUserID)
			if n <= 0 {
				return fmt.Errorf("convert user id %q to uint64", rawUserID)
			}

			token = Token{
				ID:     string(k),
				UserID: userID,
			}
			return nil
		})
	})
	return token, err
}

func TokenByID(id string) MatchFunc {
	bid := []byte(id)
	return ByKey(bid)
}
