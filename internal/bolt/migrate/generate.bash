#!/usr/bin/env bash

dir="$(dirname "${BASH_SOURCE[0]}")"
name="$1"
time="$(date -u +"%FT%R:%S")"

_usage() {
	echo "Usage: $(basename "${BASH_SOURCE[0]}") <name>"
}

generate() {
	cat << EOF
package migrate

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
)

func init() {
	v, err := time.Parse(VersionLayout, "$time")
	if err != nil {
		panic(fmt.Errorf("parse generated migration version %q: %w", "$time", err))
	}

	register(v, func(db storm.Node) error {
		panic("Not implemented.")
	})
}
EOF
}

if [[ -z "$name" ]]; then
	_usage
	exit 2
fi

file="$dir/${time}_${name}.go"

generate > "$file"
echo "$file"