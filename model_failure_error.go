package sips

type FailureError struct {

	// Mandatory string identifying the type of error
	Reason string `json:"reason"`

	// Optional, longer description of the error; may include UUID of transaction for support, links to documentation etc
	Details string `json:"details,omitempty"`
}
