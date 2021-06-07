package main

import (
	"context"
	"errors"
	"time"

	"github.com/DeedleFake/sips/dbs"
	"github.com/DeedleFake/sips/internal/log"
	"github.com/DeedleFake/sips/ipfsapi"
	"github.com/asdine/storm"
)

type JobQueue struct {
	IPFS *ipfsapi.Client
	DB   *storm.DB

	cancel context.CancelFunc
	done   chan struct{}
}

func (jq *JobQueue) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	jq.cancel = cancel
	jq.done = make(chan struct{})

	go jq.loop(ctx)
}

func (jq *JobQueue) Stop() {
	jq.cancel()
	<-jq.done
}

func (jq *JobQueue) loop(ctx context.Context) {
	defer jq.cancel()
	defer close(jq.done)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var job dbs.Job
		err := jq.DB.Select().OrderBy("Created").First(&job)
		if err != nil {
			if !errors.Is(err, storm.ErrNotFound) {
				log.Errorf("get next job from queue: %v", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		var pin dbs.Pin
		err = jq.DB.One("ID", job.Pin, &pin)
		if err != nil {
			log.Errorf("get pin for job %v: %v", job.ID, err)
			continue
		}

		switch job.Mode {
		case dbs.ModeAdd:
			jq.modeAdd(ctx, job, pin)
		case dbs.ModeRm:
			jq.modeRm(ctx, job, pin)
		}
	}
}

func (jq *JobQueue) modeAdd(ctx context.Context, job dbs.Job, pin dbs.Pin) {
	panic("Not implemented.")
}

func (jq *JobQueue) modeRm(ctx context.Context, job dbs.Job, pin dbs.Pin) {
	panic("Not implemented.")
}
