package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/ent"
	"github.com/DeedleFake/sips/ent/pin"
	"github.com/DeedleFake/sips/ent/token"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
)

func auth(ctx context.Context, db *ent.Tx) (u *ent.User, err error) {
	tokstr, _ := sips.Token(ctx)

	tok, err := db.Token.Query().
		WithUser().
		Where(token.Token(tokstr)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("find token %q: %w", tokstr, err)
	}

	return tok.Edges.User, nil
}

type PinHandler struct {
	Queue *PinQueue
	IPFS  *ipfsapi.Client
	DB    *ent.Client
}

func (h PinHandler) Pins(ctx context.Context, query sips.PinQuery) ([]sips.PinStatus, error) {
	tx, err := h.DB.Tx(ctx)
	if err != nil {
		return nil, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := auth(ctx, tx)
	if err != nil {
		return nil, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	q := u.QueryPins().
		Order(ent.Desc(pin.FieldCreateTime)).
		Limit(query.Limit)
	if len(query.Status) > 0 {
		q = q.Where(pin.StatusIn(query.Status...))
	}
	if len(query.CID) > 0 {
		q = q.Where(pin.CIDIn(query.CID...))
	}
	if !query.Before.IsZero() {
		q = q.Where(pin.CreateTimeLT(query.Before))
	}
	if !query.After.IsZero() {
		q = q.Where(pin.CreateTimeGT(query.After))
	}
	pins, err := q.All(ctx)
	if err != nil {
		return nil, log.Errorf("query pins: %w", err)
	}

	if query.Name != "" {
		// TODO: Handle this in the query.
		for i := range pins {
			if !query.Match.Match(pins[i].Name, query.Name) {
				pins = append(pins[:i], pins[i+1:]...)
				i--
			}
		}
	}

	statuses := make([]sips.PinStatus, len(pins))
	for i, pin := range pins {
		statuses[i] = sips.PinStatus{
			RequestID: strconv.FormatInt(int64(pin.ID), 16),
			Status:    pin.Status,
			Created:   pin.CreateTime,
			Pin: sips.Pin{
				CID:     pin.CID,
				Name:    pin.Name,
				Origins: pin.Origins,
			},
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, log.Errorf("commit transaction: %w", err)
	}

	return statuses, nil
}

func (h PinHandler) AddPin(ctx context.Context, pin sips.Pin) (sips.PinStatus, error) {
	tx, err := h.DB.Tx(ctx)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := auth(ctx, tx)
	if err != nil {
		return sips.PinStatus{}, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	dbpin, err := h.DB.Pin.Create().
		SetUser(u).
		SetCID(pin.CID).
		SetName(pin.Name).
		SetOrigins(pin.Origins).
		Save(ctx)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("create pin: %w", err)
	}

	select {
	case <-ctx.Done():
		return sips.PinStatus{}, log.Errorf("queue add %q: %w", pin.CID, ctx.Err())
	case h.Queue.Add() <- dbpin:
	}

	id, err := h.IPFS.ID(ctx)
	if err != nil {
		log.Errorf("get IPFS ID: %w", err)
		// Purposefully don't return here.
	}

	err = tx.Commit()
	if err != nil {
		return sips.PinStatus{}, log.Errorf("commit transaction: %w", err)
	}

	return sips.PinStatus{
		RequestID: strconv.FormatInt(int64(dbpin.ID), 16),
		Status:    dbpin.Status,
		Created:   dbpin.CreateTime,
		Delegates: id.Addresses,
		Pin:       pin,
	}, nil
}

func (h PinHandler) GetPin(ctx context.Context, requestID string) (sips.PinStatus, error) {
	pinID, err := strconv.ParseInt(requestID, 16, 64)
	if err != nil {
		return sips.PinStatus{}, BadRequest(log.Errorf("parse request ID %q: %w", requestID, err))
	}

	tx, err := h.DB.Tx(ctx)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := auth(ctx, tx)
	if err != nil {
		return sips.PinStatus{}, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	pin, err := u.QueryPins().
		Where(pin.ID(int(pinID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return sips.PinStatus{}, NotFound(log.Errorf("query pin %q: %w", requestID, err))
		}
		return sips.PinStatus{}, log.Errorf("query pin %q: %w", requestID, err)
	}

	err = tx.Commit()
	if err != nil {
		return sips.PinStatus{}, log.Errorf("commit transaction: %w", err)
	}

	return sips.PinStatus{
		RequestID: requestID,
		Status:    pin.Status,
		Created:   pin.CreateTime,
		Pin: sips.Pin{
			CID:  pin.CID,
			Name: pin.Name,
		},
	}, nil
}

func (h PinHandler) UpdatePin(ctx context.Context, requestID string, spin sips.Pin) (sips.PinStatus, error) {
	pinID, err := strconv.ParseInt(requestID, 16, 64)
	if err != nil {
		return sips.PinStatus{}, BadRequest(log.Errorf("parse request ID %q: %w", requestID, err))
	}

	tx, err := h.DB.Tx(ctx)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := auth(ctx, tx)
	if err != nil {
		return sips.PinStatus{}, Unauthorized(log.Errorf("authenticate: %w", err))
	}

	oldpin, err := u.QueryPins().
		Where(pin.ID(int(pinID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return sips.PinStatus{}, NotFound(log.Errorf("query pin %q: %w", requestID, err))
		}
		return sips.PinStatus{}, log.Errorf("query pin %q: %w", requestID, err)
	}

	newpin, err := tx.Pin.UpdateOne(oldpin).
		SetStatus(sips.Queued).
		SetCID(spin.CID).
		SetName(spin.Name).
		SetOrigins(spin.Origins).
		Save(ctx)
	if err != nil {
		return sips.PinStatus{}, log.Errorf("update pin %q: %w", requestID, err)
	}

	select {
	case <-ctx.Done():
		return sips.PinStatus{}, log.Errorf("queue update %q: %w", requestID, ctx.Err())
	case h.Queue.Update() <- [2]*ent.Pin{oldpin, newpin}:
	}

	id, err := h.IPFS.ID(ctx)
	if err != nil {
		log.Errorf("get IPFS ID: %w", err)
		// Purposefully don't return here.
	}

	err = tx.Commit()
	if err != nil {
		return sips.PinStatus{}, log.Errorf("commit transaction: %w", err)
	}

	return sips.PinStatus{
		RequestID: requestID,
		Status:    newpin.Status,
		Created:   newpin.CreateTime,
		Delegates: id.Addresses,
		Pin: sips.Pin{
			CID:  newpin.CID,
			Name: newpin.Name,
		},
	}, nil
}

func (h PinHandler) DeletePin(ctx context.Context, requestID string) error {
	pinID, err := strconv.ParseInt(requestID, 16, 64)
	if err != nil {
		return BadRequest(log.Errorf("parse request ID %q: %w", requestID, err))
	}

	tx, err := h.DB.Tx(ctx)
	if err != nil {
		return log.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := auth(ctx, tx)
	if err != nil {
		return Unauthorized(log.Errorf("authenticate: %w", err))
	}

	pin, err := u.QueryPins().
		Where(pin.ID(int(pinID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NotFound(log.Errorf("query pin %q: %w", requestID, err))
		}
		return log.Errorf("query pin %q: %w", requestID, err)
	}

	select {
	case <-ctx.Done():
		return log.Errorf("queue delete %q: %w", requestID, ctx.Err())
	case h.Queue.Delete() <- pin:
	}

	err = tx.Commit()
	if err != nil {
		return log.Errorf("commit transaction: %w", err)
	}

	return nil
}
