package sips

import (
	"time"
)

// PinStatus - Pin object with status
type PinStatus struct {

	// Globally unique identifier of the pin request; can be used to check the status of ongoing pinning, or pin removal
	Requestid string `json:"requestid"`

	Status Status `json:"status"`

	// Immutable timestamp indicating when a pin request entered a pinning service; can be used for filtering results and pagination
	Created time.Time `json:"created"`

	Pin Pin `json:"pin"`

	// List of multiaddrs designated by pinning service for transferring any new data from external peers
	Delegates []string `json:"delegates"`

	// Optional info for PinStatus response
	Info map[string]string `json:"info,omitempty"`
}
