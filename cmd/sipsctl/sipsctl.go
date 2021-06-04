package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/DeedleFake/sips/cmd/sipsctl/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
