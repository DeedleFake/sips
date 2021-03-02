package sips

// PinResults - Response used for listing pin objects matching request
type PinResults struct {

	// The total number of pin objects that exist for passed query filters
	Count int32 `json:"count"`

	// An array of PinStatus results
	Results []PinStatus `json:"results"`
}
