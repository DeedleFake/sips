package sips

// Pin - Pin object
type Pin struct {

	// Content Identifier (CID) to be pinned recursively
	Cid string `json:"cid"`

	// Optional name for pinned data; can be used for lookups later
	Name string `json:"name,omitempty"`

	// Optional list of multiaddrs known to provide the data
	Origins []string `json:"origins,omitempty"`

	// Optional metadata for pin object
	Meta map[string]string `json:"meta,omitempty"`
}
