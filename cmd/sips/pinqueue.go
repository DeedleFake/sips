package main

import (
	"context"
	"sync/atomic"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/ent"
	"github.com/DeedleFake/sips/ent/pin"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
)

// PinQueue handles queued pin requests, synchronizing them to both
// the database and IPFS.
type PinQueue struct {
	running uint32
	cancel  context.CancelFunc
	done    chan struct{}

	add    chan *ent.Pin
	update chan [2]*ent.Pin
	del    chan *ent.Pin

	IPFS *ipfsapi.Client
	DB   *ent.Client
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

	q.add = make(chan *ent.Pin)
	q.update = make(chan [2]*ent.Pin)
	q.del = make(chan *ent.Pin)

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
func (q *PinQueue) Add() chan<- *ent.Pin {
	return q.add
}

// Update returns a channel to which pairs of pins should be sent
// where the first pin is to be updated to the second one.
func (q *PinQueue) Update() chan<- [2]*ent.Pin {
	return q.update
}

// Delete returns a channel to which pins that are to be deleted
// should be sent.
func (q *PinQueue) Delete() chan<- *ent.Pin {
	return q.del
}

func (q *PinQueue) queueExisting(ctx context.Context) {
	tx, err := q.DB.Tx(ctx)
	if err != nil {
		log.Errorf("begin transaction for existing queued pins: %w", err)
		return
	}
	defer tx.Rollback()

	pins, err := tx.Pin.Query().
		Where(pin.StatusIn(sips.Queued, sips.Pinning)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Infof("no existing queued or in-progress pins")
			return
		}

		log.Errorf("query existing queued pins: %w", err)
		return
	}

	for _, pin := range pins {
		q.add <- pin
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("commit transaction for existing queued pins: %w", err)
		return
	}
}

func (q *PinQueue) run(ctx context.Context) {
	defer close(q.done)
	defer q.unsetRunning()

	add := q.add
	update := q.update
	del := q.del

	var stopping bool
	jobs := make(map[int]context.CancelFunc)
	jobdone := make(chan int)
	jobctx := func(id int) context.Context {
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

func (q *PinQueue) addPin(ctx context.Context, pin *ent.Pin) {
	tx, err := q.DB.Tx(ctx)
	if err != nil {
		log.Errorf("begin transaction for pin %d: %w", pin.ID, err)
		return
	}
	defer tx.Rollback()

	q.connect(ctx, pin.Origins)

	switch pin.Status {
	case "", sips.Queued:
		pin, err = tx.Pin.UpdateOne(pin).
			SetStatus(sips.Pinning).
			Save(ctx)
		if err != nil {
			log.Errorf("update pin %v status to pinning: %w", pin.ID, err)
			return
		}
	}

	status := pin.Status
	defer func() {
		_, err := tx.Pin.UpdateOne(pin).
			SetStatus(status).
			Save(ctx)
		if err != nil {
			log.Errorf("update pin %v status to %v: %w", pin.ID, pin.Status, err)
			return
		}

		err = tx.Commit()
		if err != nil {
			log.Errorf("commit transaction for pin %v: %w", pin.ID, err)
			return
		}
	}()

	progress, err := q.IPFS.PinAddProgress(ctx, pin.CID)
	if err != nil {
		log.Errorf("pin %v to IPFS: %w", pin.CID, err)
		status = sips.Failed
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case progress, closed := <-progress:
			if closed {
				log.Infof("Pinned %v as %q (%v)", pin.CID, pin.Name, pin.ID)
				status = sips.Pinned
				return
			}

			if progress.Err != nil {
				log.Errorf("pin %v to IPFS: %w", pin.CID, progress.Err)
				status = sips.Failed
				return
			}
		}
	}
}

func (q *PinQueue) updatePin(ctx context.Context, from, to *ent.Pin) {
	tx, err := q.DB.Tx(ctx)
	if err != nil {
		log.Errorf("begin transaction for pin %d: %w", from.ID, err)
		return
	}
	defer tx.Rollback()

	q.connect(ctx, to.Origins)

	if to.Status == sips.Queued {
		to, err = tx.Pin.UpdateOne(to).
			SetStatus(sips.Pinning).
			Save(ctx)
		if err != nil {
			log.Errorf("update pin %v status from queued to pinning: %w", to.ID, err)
			return
		}
	}

	status := to.Status
	defer func() {
		_, err := tx.Pin.UpdateOne(to).
			SetStatus(status).
			Save(ctx)
		if err != nil {
			log.Errorf("update pin %v status to %v: %w", to.ID, to.Status, err)
			return
		}

		err = tx.Commit()
		if err != nil {
			log.Errorf("commit transaction for pin %v: %w", to.ID, err)
			return
		}
	}()

	// TODO: Unpin updated pins manually if nothing else has pinned them.
	_, err = q.IPFS.PinUpdate(ctx, from.CID, to.CID, false)
	if err != nil {
		log.Errorf("update pin %v to %v: %w", from.ID, to.CID, err)
		status = sips.Failed
		return
	}
	log.Infof("Pin %v updated from %v to %v.", to.ID, from.CID, to.CID)

	status = sips.Pinned
}

func (q *PinQueue) deletePin(ctx context.Context, pin *ent.Pin) {
	tx, err := q.DB.Tx(ctx)
	if err != nil {
		log.Errorf("begin transaction for pin %d: %w", pin.ID, err)
		return
	}
	defer tx.Rollback()

	_, err = q.IPFS.PinRm(ctx, pin.CID)
	if err != nil {
		log.Errorf("remove pin %v from IPFS: %w", pin.CID, err)
		return
	}
	log.Infof("Pin %v (%q, %v) deleted.", pin.ID, pin.Name, pin.CID)

	err = tx.Pin.DeleteOneID(pin.ID).Exec(ctx)
	if err != nil {
		log.Errorf("delete pin %v from database: %w", pin.ID, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("commit transaction for pin %v: %w", pin.ID, err)
		return
	}
}
