package sips

import (
	"fmt"
	"strings"
	"time"
)

// PinQuery provides query info for finding pinning requests. Any
// fields which are zero values, or have length zero in the case of
// slices, should be considered as though they are not present.
type PinQuery struct {
	// CID is a list CIDs that are being searched for.
	CID []string

	// Name is the name of the desired pinning request.
	Name string

	// Match is the strategy to use for finding the name.
	Match TextMatchingStrategy

	// Status is a list of statuses used to filter the returned
	// requests.
	Status []RequestStatus

	// Before and After indicate creation time filters.
	Before, After time.Time

	// Limit is the maximum number of results that should be returned.
	// If the handler returns more than this, the list will be truncated
	// to match the value of this field.
	Limit int

	// Meta is used to filter against the Meta field of the returned
	// requests. Note that this is simply the result of unmarshalling
	// json using the encoding/json package, so it will likely be a
	// map[string]interface{}, but it might not be, either. It is up to
	// the handler to deal with this as it sees fit.
	Meta interface{}
}

func defaultPinQuery() PinQuery {
	return PinQuery{
		Match: Exact,
		Limit: 10,
	}
}

// TextMatchingStrategy indicates a strategy to use for matching one
// string against another.
type TextMatchingStrategy string

const (
	Exact    TextMatchingStrategy = "exact"
	IExact   TextMatchingStrategy = "iexact"
	Partial  TextMatchingStrategy = "partial"
	IPartial TextMatchingStrategy = "ipartial"
)

func (tms TextMatchingStrategy) valid() bool {
	switch tms {
	case Exact, IExact, Partial, IPartial:
		return true
	default:
		return false
	}
}

// Match performs a match using the specified strategy. It will panic
// if the strategy provided is invalid.
func (tms TextMatchingStrategy) Match(haystack, needle string) bool {
	switch tms {
	case Exact:
		return needle == haystack

	case IExact:
		return strings.ToLower(needle) == strings.ToLower(haystack)

	case Partial:
		return strings.Contains(haystack, needle)

	case IPartial:
		return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))

	default:
		panic(fmt.Errorf("invalid text matching strategy: %q", tms))
	}
}
