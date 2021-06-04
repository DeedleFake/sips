package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/dbutil"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/DeedleFake/sips/ipfsapi"
	"go.etcd.io/bbolt"
)

var (
	ErrNoToken = errors.New("no token")
)

type PinHandler struct {
	IPFS *ipfsapi.Client
	DB   *bbolt.DB
}

func (h PinHandler) Pins(ctx context.Context, query sips.PinQuery) ([]sips.PinStatus, error) {
	//tokID, ok := sips.Token(ctx)
	//if !ok {
	//	return nil, ErrNoToken
	//}

	panic("Not implemented.")
}

func (h PinHandler) AddPin(ctx context.Context, pin sips.Pin) (sips.PinStatus, error) {
	tokID, ok := sips.Token(ctx)
	if !ok {
		return sips.PinStatus{}, ErrNoToken
	}
	tok, err := dbutil.GetToken(h.DB, dbutil.TokenByID(tokID))
	if err != nil {
		return sips.PinStatus{}, AuthError{
			Token: tokID,
			Err:   err,
		}
	}

	log.Infof("authenticated user %v with token %q", tok.UserID, tok.ID)

	panic("Not implemented.")
}

func (h PinHandler) GetPin(ctx context.Context, requestID string) (sips.PinStatus, error) {
	panic("Not implemented.")
}

func (h PinHandler) UpdatePin(ctx context.Context, requestID string, pin sips.Pin) (sips.PinStatus, error) {
	panic("Not implemented.")
}

func (h PinHandler) DeletePin(ctx context.Context, requestID string) error {
	panic("Not implemented.")
}

type AuthError struct {
	Token string
	Err   error
}

func (err AuthError) Error() string {
	return fmt.Sprintf("authenticate token %q: %v", err.Token, err.Err)
}

func (err AuthError) Unwrap() error {
	return err.Err
}
