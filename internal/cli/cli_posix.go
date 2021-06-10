// +build linux unix darwin netbsd openbsd freebsd

package cli

import (
	"os"

	"golang.org/x/sys/unix"
)

var Signals = []os.Signal{
	os.Interrupt,
	unix.SIGTERM,
}
