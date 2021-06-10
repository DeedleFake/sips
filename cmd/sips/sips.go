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
	"github.com/DeedleFake/sips/dbs"
	"github.com/DeedleFake/sips/internal/cli"
	"github.com/DeedleFake/sips/internal/ipfsapi"
	"github.com/DeedleFake/sips/internal/log"
)

func run(ctx context.Context) error {
	addr := flag.String("addr", ":8080", "address to serve HTTP on")
	api := flag.String("api", "http://127.0.0.1:5001", "IPFS API to contact")
	rawdbpath := flag.String("db", "$CONFIG/sips/database.db", "path to database ($CONFIG will be replaced with user config dir path)")
	flag.Parse()

	dbpath, configDirUsed, err := cli.ExpandConfig(*rawdbpath)
	if err != nil {
		return err
	}

	ipfs := ipfsapi.NewClient(
		ipfsapi.WithBaseURL(*api),
		ipfsapi.WithHTTPClient(&http.Client{
			Timeout: 10 * time.Second, // TODO: Let the user configure this.
		}),
	)

	if configDirUsed {
		err := os.MkdirAll(filepath.Dir(dbpath), 0770)
		if err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
	}
	db, err := dbs.Open(dbpath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	log.Infof("Database opened at %q", dbpath)

	ph := PinHandler{
		IPFS: ipfs,
		DB:   db,
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := run(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
