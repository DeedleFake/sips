package migrate

import (
	"sort"
	"time"
)

var migrations []migrationRegistration

type migrationRegistration struct {
	V time.Time
	M migration
}

func register(version time.Time, migration Migration) {
	i := sort.Search(len(migrations), func(i int) bool {
		return version.After(migrations[i].V)
	})
	migrations = append(
		migrations[:i],
		append(
			[]migrationRegistration{{V: version, M: migration}},
			migrations[i:]...,
		)...,
	)
}
