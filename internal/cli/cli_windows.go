package cli

import (
	"os"

	"golang.org/x/sys/windows"
)

var Signals = []os.Signal{
	os.Interrupt,
	windows.SIGTERM,
}
