package sips

// TextMatchingStrategy : Alternative text matching strategy
type TextMatchingStrategy string

// List of TextMatchingStrategy
const (
	EXACT    TextMatchingStrategy = "exact"
	IEXACT   TextMatchingStrategy = "iexact"
	PARTIAL  TextMatchingStrategy = "partial"
	IPARTIAL TextMatchingStrategy = "ipartial"
)
