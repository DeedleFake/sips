package sips

import (
	"context"
	"net/http"
	"time"
)

// PinsApiRouter defines the required methods for binding the api requests to a responses for the PinsApi
// The PinsApiRouter implementation should parse necessary information from the http request,
// pass the data to a PinsApiServicer to perform the required actions, then write the service results to the http response.
type PinsApiRouter interface {
	PinsGet(http.ResponseWriter, *http.Request)
	PinsPost(http.ResponseWriter, *http.Request)
	PinsRequestidDelete(http.ResponseWriter, *http.Request)
	PinsRequestidGet(http.ResponseWriter, *http.Request)
	PinsRequestidPost(http.ResponseWriter, *http.Request)
}

// PinsApiServicer defines the api actions for the PinsApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type PinsApiServicer interface {
	PinsGet(context.Context, []string, string, TextMatchingStrategy, []Status, time.Time, time.Time, int32, map[string]string) (ImplResponse, error)
	PinsPost(context.Context, Pin) (ImplResponse, error)
	PinsRequestidDelete(context.Context, string) (ImplResponse, error)
	PinsRequestidGet(context.Context, string) (ImplResponse, error)
	PinsRequestidPost(context.Context, string, Pin) (ImplResponse, error)
}
