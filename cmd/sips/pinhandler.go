package main

import (
	"context"
	"errors"
	"strconv"
	"time"

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

	tx, err := h.DB.Begin(true)
	if err != nil {
		log.Errorf("begin transaction: %v", err)
		return sips.PinStatus{}, err
	}
	defer tx.Rollback()

	var tok dbs.Token
	err = tx.One("ID", tokID, &tok)
	if err != nil {
		log.Errorf("find token %q: %v", tokID, err)
		return sips.PinStatus{}, err
	}

	var user dbs.User
	err = tx.One("ID", tok.User, &user)
	if err != nil {
		log.Errorf("find user %q: %v", tok.User, err)
		return sips.PinStatus{}, err
	}

	dbpin := dbs.Pin{
		Created: time.Now(),
		User:    user.ID,
		Name:    pin.Name,
		CID:     pin.CID,
	}
	err = tx.Save(&dbpin)
	if err != nil {
		log.Errorf("save pin %q: %v", pin.CID, err)
		return sips.PinStatus{}, err
	}

	if len(pin.Origins) != 0 {
		for _, origin := range pin.Origins {
			go h.IPFS.SwarmConnect(origin)
		}
	}

	_, err = h.IPFS.PinAdd(pin.CID)
	if err != nil {
		log.Errorf("add pin %v: %v", pin.CID, err)
		return sips.PinStatus{}, err
	}

	id, err := h.IPFS.ID()
	if err != nil {
		log.Errorf("get IPFS id: %v", err)
		// Purposefully don't return here.
	}

	return sips.PinStatus{
		RequestID: strconv.FormatUint(dbpin.ID, 16),
		Status:    sips.Pinning,
		Created:   dbpin.Created,
		Delegates: id.Addresses,
		Pin:       pin,
	}, tx.Commit()
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
