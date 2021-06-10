package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const flags = log.Ldate | log.Ltime | log.LUTC | log.Lmsgprefix

var (
	info  = log.New(os.Stderr, "[INFO] ", flags)
	err   = log.New(os.Stderr, "[ERROR] ", flags)
	fatal = log.New(os.Stderr, "[FATAL] ", flags)

	packageDir string
)

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get package directory information")
	}
	packageDir = filepath.Join(filepath.Dir(file), "..", "..")
}

func getLocation() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown location"
	}
	file = strings.TrimPrefix(file, packageDir)[1:]
	return fmt.Sprintf("%v:%v", file, line)
}

// Infof logs an informational message.
func Infof(str string, args ...interface{}) {
	loc := getLocation()
	msg := fmt.Sprintf(str, args...)
	info.Printf("(%v) %v", loc, msg)
}

// Errorf logs an error. As a special case, it returns a new error
// constructed from its arguments via fmt.Errorf.
func Errorf(str string, args ...interface{}) error {
	loc := getLocation()
	nerr := fmt.Errorf(str, args...)
	err.Printf("(%v) %v", loc, nerr)
	return nerr
}

// Fatalf logs a fatal error and immediately exits.
func Fatalf(str string, args ...interface{}) {
	loc := getLocation()
	fatal.Fatalf("(%v) %v", loc, fmt.Sprintf(str, args...))
}
