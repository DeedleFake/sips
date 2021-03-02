package sips

// Status : Status a pin object can have at a pinning service
type Status string

// List of Status
const (
	QUEUED  Status = "queued"
	PINNING Status = "pinning"
	PINNED  Status = "pinned"
	FAILED  Status = "failed"
)
