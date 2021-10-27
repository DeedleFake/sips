// sips is the primary implementation of a simple pinning service daemon.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/db"
	"github.com/DeedleFake/sips/internal/cli"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
)

func run(ctx context.Context) error {
	addr := flag.String("addr", ":8080", "address to serve HTTP on")
	api := flag.String("api", "http://127.0.0.1:5001", "IPFS API to contact")
	apitimeout := flag.Duration("apitimeout", 30*time.Second, "timeout for requests to the IPFS API")
	rawdbpath := flag.String("db", "host=/var/run/postgresql dbname=sips", "path to database ($CONFIG will be replaced with user config dir path)")
	domigration := flag.Bool("migrate", true, "perform a database migration upon starting")
	flag.Parse()

	dbpath, configDirUsed, err := cli.ExpandConfig(*rawdbpath)
	if err != nil {
		return err
	}

	ipfs := ipfsapi.NewClient(
		ipfsapi.WithBaseURL(*api),
		ipfsapi.WithHTTPClient(&http.Client{
			Timeout: *apitimeout,
		}),
	)

	if configDirUsed {
		err := os.MkdirAll(filepath.Dir(dbpath), 0770)
		if err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
	}
	entc, err := db.Open("postgres", dbpath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer entc.Close()
	log.Infof("Database opened at %q", dbpath)

	if *domigration {
		log.Infof("Running migrations...")
		err = entc.Schema.Create(ctx)
		if err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}

	queue := PinQueue{
		IPFS: ipfs,
		DB:   entc,
	}
	queue.Start(ctx)
	defer queue.Stop()

	ph := PinHandler{
		Queue: &queue,
		IPFS:  ipfs,
		DB:    entc,
	}

	server := http.Server{
		Addr:    *addr,
		Handler: sips.Handler(&ph),
		BaseContext: func(lis net.Listener) context.Context {
			return ctx
		},
	}

	shutdown := make(chan error, 1)
	go func() {
		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Infof("Exiting...")
		shutdown <- server.Shutdown(sctx)
	}()

	log.Infof("Starting server...")
	err = server.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("start server: %w", err)
		}

		err = <-shutdown
		if err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), cli.Signals...)
	defer cancel()

	err := run(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
