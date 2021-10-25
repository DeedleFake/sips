package main

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/DeedleFake/sips"
	dbs "github.com/DeedleFake/sips/internal/bolt"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/asdine/storm"
	sq "github.com/asdine/storm/q"
)

// PinQueue handles queued pin requests, synchronizing them to both
// the database and IPFS.
type PinQueue struct {
	running uint32
	cancel  context.CancelFunc
	done    chan struct{}

	add    chan dbs.Pin
	update chan [2]dbs.Pin
	del    chan dbs.Pin

	IPFS *ipfsapi.Client
	DB   *storm.DB
}

func (q *PinQueue) setRunning() bool {
	return atomic.CompareAndSwapUint32(&q.running, 0, 1)
}

func (q *PinQueue) unsetRunning() {
	atomic.StoreUint32(&q.running, 0)
}

// Start starts the queue. No other methods should be called before
// this one returns, and calls to this method while the queue is
// running will panic.
func (q *PinQueue) Start(ctx context.Context) {
	if !q.setRunning() {
		panic("already running")
	}

	ctx, q.cancel = context.WithCancel(ctx)
	q.done = make(chan struct{})

	q.add = make(chan dbs.Pin)
	q.update = make(chan [2]dbs.Pin)
	q.del = make(chan dbs.Pin)

	go q.run(ctx)
	q.queueExisting(ctx)
}

// Stop stops a running queue. It does not return until the queue has
// completely flushed all of its jobs.
func (q *PinQueue) Stop() {
	q.cancel()
	<-q.done
}

// Add returns a channel to which pins that are to be added should be
// sent.
func (q *PinQueue) Add() chan<- dbs.Pin {
	return q.add
}

// Update returns a channel to which pairs of pins should be sent
// where the first pin is to be updated to the second one.
func (q *PinQueue) Update() chan<- [2]dbs.Pin {
	return q.update
}

// Delete returns a channel to which pins that are to be deleted
// should be sent.
func (q *PinQueue) Delete() chan<- dbs.Pin {
	return q.del
}

func (q *PinQueue) queueExisting(ctx context.Context) {
	tx, err := q.DB.Begin(false)
	if err != nil {
		log.Errorf("begin transaction for existing queued pins: %w", err)
		return
	}
	defer tx.Rollback()

	var pins []dbs.Pin
	err = tx.Select(sq.In("Status", []sips.RequestStatus{sips.Queued, sips.Pinning})).Find(&pins)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			log.Infof("No existing queued or in-progress pins.")
			return
		}

		log.Errorf("find existing queued pins: %w", err)
		return
	}

	for _, pin := range pins {
		q.add <- pin
	}
}

func (q *PinQueue) run(ctx context.Context) {
	defer close(q.done)
	defer q.unsetRunning()

	add := q.add
	update := q.update
	del := q.del

	var stopping bool
	jobs := make(map[uint64]context.CancelFunc)
	jobdone := make(chan uint64)
	jobctx := func(id uint64) context.Context {
		if cancel, ok := jobs[id]; ok {
			cancel()
			// TODO: Wait for the job to actually cancel completely?
		}

		ctx, cancel := context.WithCancel(ctx)
		jobs[id] = cancel
		return ctx
	}

	ctxdone := ctx.Done()
	for {
		select {
		case <-ctxdone:
			ctxdone = nil

			add = nil
			update = nil
			del = nil

			stopping = true
			if len(jobs) == 0 {
				return
			}

		case id := <-jobdone:
			delete(jobs, id)
			if stopping && (len(jobs) == 0) {
				return
			}

		case pin := <-add:
			sub := jobctx(pin.ID)
			go func() {
				q.addPin(sub, pin)
				jobdone <- pin.ID
			}()

		case pins := <-update:
			sub := jobctx(pins[1].ID)
			go func() {
				q.updatePin(sub, pins[0], pins[1])
				jobdone <- pins[1].ID
			}()

		case pin := <-del:
			sub := jobctx(pin.ID)
			go func() {
				q.deletePin(sub, pin)
				jobdone <- pin.ID
			}()
		}
	}
}

func (q *PinQueue) connect(ctx context.Context, origins []string) {
	for _, origin := range origins {
		go q.IPFS.SwarmConnect(ctx, origin)
	}
}

func (q *PinQueue) addPin(ctx context.Context, pin dbs.Pin) {
	q.connect(ctx, pin.Origins)

	switch pin.Status {
	case "", sips.Queued:
		pin.Status = sips.Pinning
		err := q.DB.Update(&pin)
		if err != nil {
			log.Errorf("update pin %v status to pinning: %w", pin.ID, err)
			return
		}
	}

	defer func() {
		err := q.DB.Update(&pin)
		if err != nil {
			log.Errorf("update pin %v status to %v: %w", pin.ID, pin.Status, err)
			return
		}
	}()

	progress, err := q.IPFS.PinAddProgress(ctx, pin.CID)
	if err != nil {
		log.Errorf("pin %v to IPFS: %w", pin.CID, err)
		pin.Status = sips.Failed
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case progress, closed := <-progress:
			if closed {
				log.Infof("Pinned %v as %q (%v)", pin.CID, pin.Name, pin.ID)
				pin.Status = sips.Pinned
				return
			}

			if progress.Err != nil {
				log.Errorf("pin %v to IPFS: %w", pin.CID, progress.Err)
				pin.Status = sips.Failed
				return
			}
		}
	}
}

func (q *PinQueue) updatePin(ctx context.Context, from, to dbs.Pin) {
	q.connect(ctx, to.Origins)

	if to.Status == sips.Queued {
		to.Status = sips.Pinning
		err := q.DB.Update(&to)
		if err != nil {
			log.Errorf("update pin %v status from queued to pinning: %w", to.ID, err)
			return
		}
	}

	defer func() {
		err := q.DB.Update(&to)
		if err != nil {
			log.Errorf("update pin %v status to %v: %w", to.ID, to.Status, err)
			return
		}
	}()

	// TODO: Unpin updated pins manually if nothing else has pinned them.
	_, err := q.IPFS.PinUpdate(ctx, from.CID, to.CID, false)
	if err != nil {
		log.Errorf("update pin %v to %v: %w", from.ID, to.CID, err)
		to.Status = sips.Failed
		return
	}
	log.Infof("Pin %v updated from %v to %v.", to.ID, from.CID, to.CID)

	to.Status = sips.Pinned
}

func (q *PinQueue) deletePin(ctx context.Context, pin dbs.Pin) {
	_, err := q.IPFS.PinRm(ctx, pin.CID)
	if err != nil {
		log.Errorf("remove pin %v from IPFS: %w", pin.CID, err)
		return
	}
	log.Infof("Pin %v (%q, %v) deleted.", pin.ID, pin.Name, pin.CID)

	err = q.DB.DeleteStruct(&pin)
	if err != nil {
		log.Errorf("delete pin %v from database: %w", pin.ID, err)
		return
	}
}
