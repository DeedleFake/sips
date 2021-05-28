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
	"time"

	"github.com/DeedleFake/sips/internal/log"
)

func expand(s string, mapping func(string) (string, error)) (ex string, exerr error) {
	defer func() {
		r := recover()
		switch r := r.(type) {
		case error:
			exerr = r
		case nil:
			return
		default:
			panic(r)
		}
	}()

	ex = os.Expand(s, func(env string) string {
		str, err := mapping(env)
		if err != nil {
			panic(err)
		}
		return str
	})
	return ex, exerr
}

func run(ctx context.Context) error {
	addr := flag.String("addr", ":8080", "address to serve HTTP on")
	api := flag.String("api", "http://127.0.0.1:5001", "IPFS API to contact")
	rawdbpath := flag.String("db", "$CONFIG/sips/pins.db", "path to database ($CONFIG will be replaced with user config dir path)")
	flag.Parse()

	dbpath, err := expand(*rawdbpath, func(env string) (string, error) {
		switch env {
		case "CONFIG":
			cfgdir, err := os.UserConfigDir()
			if err != nil {
				return "", fmt.Errorf("get user config directory: %w", err)
			}
			return cfgdir, nil
		default:
			return "", fmt.Errorf("unexpected variable: %q", env)
		}
	})
	if err != nil {
		return err
	}

	// TODO: Implement a handler.
	_ = api
	_ = dbpath

	server := http.Server{
		Addr: *addr,
		//Handler: handler,
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
