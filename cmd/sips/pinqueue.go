package main

import (
	"context"
	"sync/atomic"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/dbs"
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
			go q.addPin(sub, jobdone, pin)

		case pins := <-update:
			sub := jobctx(pins[1].ID)
			go q.updatePin(sub, jobdone, pins[0], pins[1])

		case pin := <-del:
			sub := jobctx(pin.ID)
			go q.deletePin(sub, jobdone, pin)
		}
	}
}

func (q *PinQueue) addPin(ctx context.Context, done chan<- uint64, pin dbs.Pin) {
	defer func() {
		done <- pin.ID
	}()

	panic("Not implemented.")
}

func (q *PinQueue) updatePin(ctx context.Context, done chan<- uint64, from, to dbs.Pin) {
	defer func() {
		// from and to should always have the same ID, so the choice here
		// is arbitrary.
		done <- to.ID
	}()

	panic("Not implemented.")
}

func (q *PinQueue) deletePin(ctx context.Context, done chan<- uint64, pin dbs.Pin) {
	defer func() {
		done <- pin.ID
	}()

	panic("Not implemented.")
}