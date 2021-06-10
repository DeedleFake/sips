// sipsctl is a simple utility for administrating the database used by SIPS.
package main

import (
	"context"
	"os/signal"

	"github.com/DeedleFake/sips/cmd/sipsctl/cmd"
	"github.com/DeedleFake/sips/internal/cli"
	"github.com/DeedleFake/sips/internal/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), cli.Signals...)
	defer cancel()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
