package dbutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

const (
	userBucket = "USERS"

	tokenBucket  = "TOKENS"
	tokenUserKey = "USER"
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

func GetUserForToken(db *bbolt.DB, tok string) (user string, err error) {
	btok := []byte(tok)
	err = db.View(func(tx *bbolt.Tx) error {
		tokenBucket := FindPath(
			tx,
			[]byte(tokenBucket),
			btok,
		)
		if tokenBucket == nil {
			return ErrInvalidToken
		}

		u := tokenBucket.Get([]byte(tokenUserKey))
		if u == nil {
			return ErrInvalidToken
		}

		user = string(u)
		return nil
	})
	return user, err
}

type Bucketer interface {
	Bucket([]byte) *bbolt.Bucket
}

func FindPath(b Bucketer, path ...[]byte) (r *bbolt.Bucket) {
	r = b.Bucket(path[0])
	for _, name := range path[1:] {
		if r == nil {
			return nil
		}
		r = r.Bucket(name)
	}
	return r
}
