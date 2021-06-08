package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"time"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/dbs"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/asdine/storm"
)

type PinHandler struct {
	IPFS *ipfsapi.Client
	DB   *storm.DB
}

func (h PinHandler) Pins(ctx context.Context, query sips.PinQuery) ([]sips.PinStatus, error) {
	tx, err := h.DB.Begin(false)
	if err != nil {
		log.Errorf("begin transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		log.Errorf("authenticate: %v", err)
		return nil, err
	}

	selector := tx.Select(&queryMatcher{Query: query})
	if query.Limit > 0 {
		selector = selector.Limit(query.Limit)
	}

	var dbpins []dbs.Pin
	err = selector.OrderBy("Created").Find(&dbpins)
	if (err != nil) && (!errors.Is(err, storm.ErrNotFound)) {
		log.Errorf("find pins for %v: %v", user.Name, err)
		return nil, err
	}

	pins := make([]sips.PinStatus, 0, len(dbpins))
	for _, pin := range dbpins {
		pins = append(pins, sips.PinStatus{
			RequestID: strconv.FormatUint(pin.ID, 16),
			Status:    sips.Pinned, // TODO: Handle this properly.
			Created:   pin.Created,
			Pin: sips.Pin{
				CID:  pin.CID,
				Name: pin.Name,
			},
		})
	}

	return pins, nil
}

func (h PinHandler) AddPin(ctx context.Context, pin sips.Pin) (sips.PinStatus, error) {
	tx, err := h.DB.Begin(true)
	if err != nil {
		log.Errorf("begin transaction: %v", err)
		return sips.PinStatus{}, err
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		log.Errorf("authenticate: %v", err)
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
	pinID, err := strconv.ParseUint(requestID, 16, 64)
	if err != nil {
		log.Errorf("parse request ID %q: %v", requestID, err)
		return sips.PinStatus{}, err
	}

	tx, err := h.DB.Begin(false)
	if err != nil {
		log.Errorf("begin transaction: %v", err)
		return sips.PinStatus{}, err
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		log.Errorf("authenticate: %v", err)
		return sips.PinStatus{}, err
	}

	var pin dbs.Pin
	err = tx.One("ID", pinID, &pin)
	if err != nil {
		log.Errorf("find pin %v: %v", requestID, err)
		return sips.PinStatus{}, err
	}

	if pin.User != user.ID {
		log.Errorf("user %v is not authorized to see pin %v", user.Name, requestID)
		return sips.PinStatus{}, fs.ErrPermission
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
		log.Errorf("parse request ID: %q: %v", requestID, err)
		return sips.PinStatus{}, err
	}

	tx, err := h.DB.Begin(true)
	if err != nil {
		log.Errorf("begin transaction: %v", err)
		return sips.PinStatus{}, err
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		log.Errorf("authenticate: %v", err)
		return sips.PinStatus{}, err
	}

	var dbpin dbs.Pin
	err = tx.One("ID", pinID, &dbpin)
	if err != nil {
		log.Errorf("find pin %v: %v", requestID, err)
		return sips.PinStatus{}, err
	}
	oldCID := dbpin.CID

	if dbpin.User != user.ID {
		log.Errorf("user %v not allowed to update pin %v", user.Name, requestID)
		return sips.PinStatus{}, fs.ErrPermission
	}

	dbpin.Name = pin.Name
	dbpin.CID = pin.CID
	err = tx.Update(&dbpin)
	if err != nil {
		log.Errorf("update pin %v: %v", requestID, err)
		return sips.PinStatus{}, err
	}

	if len(pin.Origins) != 0 {
		for _, origin := range pin.Origins {
			go h.IPFS.SwarmConnect(origin)
		}
	}

	_, err = h.IPFS.PinUpdate(oldCID, pin.CID, false)
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
		RequestID: requestID,
		Status:    sips.Pinning,
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
		log.Errorf("parse request ID %q: %v", requestID, err)
		return err
	}

	tx, err := h.DB.Begin(true)
	if err != nil {
		log.Errorf("begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	user, err := auth(ctx, tx)
	if err != nil {
		log.Errorf("authenticate: %v", err)
		return err
	}

	var pin dbs.Pin
	err = tx.One("ID", pinID, &pin)
	if err != nil {
		log.Errorf("find pin %v: %v", requestID, err)
		return err
	}

	if pin.User != user.ID {
		log.Errorf("user %v is not authorized to delete pin %v", user.Name, requestID)
		return fs.ErrPermission // TODO: That's just not right.
	}

	err = tx.DeleteStruct(&pin)
	if err != nil {
		log.Errorf("delete pin %v: %v", requestID, err)
		return err
	}

	_, err = h.IPFS.PinRm(pin.CID)
	if err != nil {
		log.Errorf("unpin %v: %v", pin.CID, err)
		return err
	}

	return tx.Commit()
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
