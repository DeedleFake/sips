package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/DeedleFake/sips/cmd/sipsctl/cmd"
	"github.com/DeedleFake/sips/internal/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
