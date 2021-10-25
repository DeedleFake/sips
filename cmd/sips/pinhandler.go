package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/DeedleFake/sips"
	dbs "github.com/DeedleFake/sips/internal/bolt"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/asdine/storm"
)

type PinHandler struct {
	Queue *PinQueue
	IPFS  *ipfsapi.Client
	DB    *storm.DB
}

func (h PinHandler) Pins(ctx context.Context, query sips.PinQuery) ([]sips.PinStatus, error) {
	tx, err := h.DB.Begin(false)
	if err != nil {
		return nil, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		return nil, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	selector := tx.Select(&queryMatcher{Query: query})
	if query.Limit > 0 {
		selector = selector.Limit(query.Limit)
	}

	var dbpins []dbs.Pin
	err = selector.OrderBy("Created").Find(&dbpins)
	if (err != nil) && (!errors.Is(err, storm.ErrNotFound)) {
		return nil, log.Errorf("find pins for %v: %w", user.Name, err)
	}

	pins := make([]sips.PinStatus, 0, len(dbpins))
	for _, pin := range dbpins {
		pins = append(pins, sips.PinStatus{
			RequestID: strconv.FormatUint(pin.ID, 16),
			Status:    pin.Status,
			Created:   pin.Created,
			Pin: sips.Pin{
				CID:     pin.CID,
				Name:    pin.Name,
				Origins: pin.Origins,
			},
		})
	}

	return pins, nil
}

func (h PinHandler) AddPin(ctx context.Context, pin sips.Pin) (sips.PinStatus, error) {
	tx, err := h.DB.Begin(true)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		return sips.PinStatus{}, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	dbpin := dbs.Pin{
		Created: time.Now(),
		User:    user.ID,
		Status:  sips.Queued,
		Name:    pin.Name,
		CID:     pin.CID,
		Origins: pin.Origins,
	}
	err = tx.Save(&dbpin)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("save pin %q: %w", pin.CID, err)
	}

	select {
	case <-ctx.Done():
		return sips.PinStatus{}, log.Errorf("queue add %v: %w", pin.CID, err)
	case h.Queue.Add() <- dbpin:
	}

	id, err := h.IPFS.ID(ctx)
	if err != nil {
		log.Errorf("get IPFS id: %w", err)
		// Purposefully don't return here.
	}

	return sips.PinStatus{
		RequestID: strconv.FormatUint(dbpin.ID, 16),
		Status:    dbpin.Status,
		Created:   dbpin.Created,
		Delegates: id.Addresses,
		Pin:       pin,
	}, tx.Commit()
}

func (h PinHandler) GetPin(ctx context.Context, requestID string) (sips.PinStatus, error) {
	pinID, err := strconv.ParseUint(requestID, 16, 64)
	if err != nil {
		return sips.PinStatus{}, BadRequest(log.Errorf("parse request ID %q: %w", requestID, err))
	}

	tx, err := h.DB.Begin(false)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		return sips.PinStatus{}, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	var pin dbs.Pin
	err = tx.One("ID", pinID, &pin)
	if err != nil {
		err = log.Errorf("find pin %v: %w", requestID, err)
		if errors.Is(err, storm.ErrNotFound) {
			err = NotFound(err)
		}
		return sips.PinStatus{}, err
	}

	if pin.User != user.ID {
		return sips.PinStatus{}, NotFound(log.Errorf("find pin %v: %w", requestID, storm.ErrNotFound))
	}

	return sips.PinStatus{
		RequestID: requestID,
		Status:    sips.Pinned,
		Created:   pin.Created,
		Pin: sips.Pin{
			CID:  pin.CID,
			Name: pin.Name,
		},
	}, nil
}

func (h PinHandler) UpdatePin(ctx context.Context, requestID string, pin sips.Pin) (sips.PinStatus, error) {
	pinID, err := strconv.ParseUint(requestID, 16, 64)
	if err != nil {
		return sips.PinStatus{}, BadRequest(log.Errorf("parse request ID: %q: %w", requestID, err))
	}

	tx, err := h.DB.Begin(true)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		return sips.PinStatus{}, NotFound(log.Errorf("authenticate: %w", err))
	}

	var dbpin dbs.Pin
	err = tx.One("ID", pinID, &dbpin)
	if err != nil {
		err = log.Errorf("find pin %v: %w", requestID, err)
		if errors.Is(err, storm.ErrNotFound) {
			err = NotFound(err)
		}
		return sips.PinStatus{}, err
	}

	if dbpin.User != user.ID {
		return sips.PinStatus{}, NotFound(log.Errorf("find pin %v: %w", requestID, storm.ErrNotFound))
	}

	oldpin := dbpin
	dbpin.Status = sips.Queued
	dbpin.Name = pin.Name
	dbpin.CID = pin.CID
	dbpin.Origins = pin.Origins
	err = tx.Update(&dbpin)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("update pin %v: %w", requestID, err)
	}

	select {
	case <-ctx.Done():
		return sips.PinStatus{}, log.Errorf("queue update %v: %w", requestID, ctx.Err())
	case h.Queue.Update() <- [2]dbs.Pin{oldpin, dbpin}:
	}

	id, err := h.IPFS.ID(ctx)
	if err != nil {
		log.Errorf("get IPFS id: %w", err)
		// Purposefully don't return here.
	}

	return sips.PinStatus{
		RequestID: requestID,
		Status:    dbpin.Status,
		Created:   dbpin.Created,
		Delegates: id.Addresses,
		Pin: sips.Pin{
			CID:  pin.CID,
			Name: pin.Name,
		},
	}, tx.Commit()
}

func (h PinHandler) DeletePin(ctx context.Context, requestID string) error {
	pinID, err := strconv.ParseUint(requestID, 16, 64)
	if err != nil {
		return BadRequest(log.Errorf("parse request ID %q: %w", requestID, err))
	}

	tx, err := h.DB.Begin(false)
	if err != nil {
		return log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		return Unauthorized(log.Errorf("authenticate: %w", err))
	}

	var pin dbs.Pin
	err = tx.One("ID", pinID, &pin)
	if err != nil {
		err = log.Errorf("find pin %v: %w", requestID, err)
		if errors.Is(err, storm.ErrNotFound) {
			err = NotFound(err)
		}
		return err
	}

	if pin.User != user.ID {
		return NotFound(log.Errorf("find pin %v: %w", requestID, storm.ErrNotFound))
	}

	select {
	case <-ctx.Done():
		return log.Errorf("queue delete %v: %w", requestID, ctx.Err())
	case h.Queue.Delete() <- pin:
	}

	return nil
}

func auth(ctx context.Context, db storm.Node) (user dbs.User, err error) {
	tokID, _ := sips.Token(ctx)

	var tok dbs.Token
	err = db.One("ID", tokID, &tok)
	if err != nil {
		return dbs.User{}, fmt.Errorf("find token %v: %w", tokID, err)
	}

	err = db.One("ID", tok.User, &user)
	if err != nil {
		return dbs.User{}, fmt.Errorf("find user: %w", err)
	}

	return user, nil
}

type queryMatcher struct {
	Query sips.PinQuery
}

func (qm queryMatcher) Match(v interface{}) (bool, error) {
	pin, ok := v.(dbs.Pin)
	if !ok {
		return false, fmt.Errorf("expected pin, not %T", v)
	}

	if len(qm.Query.CID) != 0 {
		var found bool
		for _, cid := range qm.Query.CID {
			if cid == pin.CID {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	if qm.Query.Name != "" {
		if !qm.Query.Match.Match(pin.Name, qm.Query.Name) {
			return false, nil
		}
	}

	// TODO: Handle statuses.

	if !qm.Query.Before.IsZero() {
		if !pin.Created.After(qm.Query.Before) {
			return false, nil
		}
	}
	if !qm.Query.After.IsZero() {
		if !pin.Created.Before(qm.Query.After) {
			return false, nil
		}
	}

	return true, nil
}
