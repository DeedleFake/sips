package sips

import "time"

// RequestStatus is the status of a given pinning request.
type RequestStatus string

const (
	Queued  RequestStatus = "queued"
	Pinning RequestStatus = "pinning"
	Pinned  RequestStatus = "pinned"
	Failed  RequestStatus = "failed"
)

// PinStatus indicates the status of a pinning request and provides
// associated info.
type PinStatus struct {
	RequestID string        `json:"requestid"`
	Status    RequestStatus `json:"status"`
	Created   time.Time     `json:"created"`
	Delegates []string      `json:"delegates"`
	Info      interface{}   `json:"info,omitempty"`

	Pin Pin `json:"pin"`
}

// Pin describes a single pinned item.
type Pin struct {
	CID     string      `json:"cid"`
	Name    string      `json:"name"`
	Origins []string    `json:"origins"`
	Meta    interface{} `json:"meta,omitempty"`
}
