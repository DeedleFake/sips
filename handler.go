package sips

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type ctxKeyToken struct{}

func withToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxKeyToken{}, token)
}

// Token returns the auth token associated with the context.
func Token(ctx context.Context) (string, bool) {
	tok, ok := ctx.Value(ctxKeyToken{}).(string)
	return tok, ok
}

// PinHandler is an interface satisfied by types that can be used to
// handle pinning service requests.
type PinHandler interface {
	// Pins returns a list of pinning request statuses based on the
	// given query.
	Pins(ctx context.Context, query PinQuery) ([]PinStatus, error)

	// AddPin adds a new pin to the service's backend.
	AddPin(ctx context.Context, pin Pin) (PinStatus, error)

	// GetPin gets the status of a specific pinning request.
	GetPin(ctx context.Context, requestID string) (PinStatus, error)

	// UpdatePin replaces a pinning request's pin with new info.
	UpdatePin(ctx context.Context, requestID string, pin Pin) (PinStatus, error)

	// DeletePin removes a pinning request.
	DeletePin(ctx context.Context, requestID string) error
}

type handler struct {
	h PinHandler
}

// Handler returns a new HTTP handler that uses h to handle pinning
// service requests.
func Handler(h PinHandler) http.Handler {
	r := mux.NewRouter()

	handler := handler{h: h}
	r.Methods("GET", "OPTIONS").Path("/pins").HandlerFunc(handler.getPins)
	r.Methods("POST", "OPTIONS").Path("/pins").HandlerFunc(handler.postPins)
	r.Methods("GET", "OPTIONS").Path("/pins/{requestID}").HandlerFunc(handler.getPinByID)
	r.Methods("POST", "OPTIONS").Path("/pins/{requestID}").HandlerFunc(handler.postPinByID)
	r.Methods("DELETE", "OPTIONS").Path("/pins/{requestID}").HandlerFunc(handler.deletePinByID)
	r.Use(mux.CORSMethodMiddleware(r))

	return r
}

func (h handler) getPins(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

func (h handler) postPins(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

func (h handler) getPinByID(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

func (h handler) postPinByID(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

func (h handler) deletePinByID(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}
