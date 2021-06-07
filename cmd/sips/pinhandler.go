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
	Jobs *JobQueue
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

	job := dbs.Job{
		Created: dbpin.Created,
		Pin:     dbpin.ID,
		Mode:    dbs.ModeAdd,
	}
	err = tx.Save(&job)
	if err != nil {
		log.Errorf("save job: %q: %v", pin.CID, err)
		return sips.PinStatus{}, err
	}

	return sips.PinStatus{
		RequestID: strconv.FormatUint(job.ID, 10),
		Status:    sips.Queued,
		Created:   time.Now(),
		// TODO: Add delegates.
		Pin: pin,
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
