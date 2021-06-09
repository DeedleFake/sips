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

func Infof(str string, args ...interface{}) {
	loc := getLocation()
	info.Printf("(%v) %v", loc, fmt.Sprintf(str, args...))
}

func Errorf(str string, args ...interface{}) {
	loc := getLocation()
	err.Printf("(%v) %v", loc, fmt.Sprintf(str, args...))
}

func Fatalf(str string, args ...interface{}) {
	loc := getLocation()
	fatal.Fatalf("(%v) %v", loc, fmt.Sprintf(str, args...))
}
