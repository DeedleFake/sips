package cli

import (
	"fmt"
	"os"
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

func ExpandConfig(str string) (path string, expanded bool, err error) {
	path, err = expand(str, func(env string) (string, error) {
		switch env {
		case "CONFIG":
			expanded = true
			cfgdir, err := os.UserConfigDir()
			if err != nil {
				return "", fmt.Errorf("get user config directory: %w", err)
			}
			return cfgdir, nil

		default:
			return "", fmt.Errorf("unexpected variable: %q", env)
		}
	})
	return path, expanded, err
}
