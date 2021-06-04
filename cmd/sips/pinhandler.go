package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/dbs"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/DeedleFake/sips/ipfsapi"
	"github.com/asdine/storm"
)

var (
	ErrNoToken    = errors.New("no token")
	ErrNoSuchUser = errors.New("user doesn't exist")
)

type PinHandler struct {
	IPFS *ipfsapi.Client
	DB   *storm.DB
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

	var tok dbs.Token
	err := h.DB.One("ID", tokID, &tok)
	if err != nil {
		log.Errorf("find token %q: %v", tokID, err)
		return sips.PinStatus{}, err
	}

	var user dbs.User
	err = h.DB.One("User", tok.User, &user)
	if err != nil {
		log.Errorf("find user %q: %v", tok.User, err)
		return sips.PinStatus{}, err
	}

	log.Infof("authenticated user %q with token %q", user.Name, tok.ID)

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
