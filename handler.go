package sips

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func tokenFromRequest(req *http.Request) (string, bool) {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) < 2 {
		return "", false
	}
	if parts[0] != "Bearer" {
		return "", false
	}

	return parts[1], true
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
// service requests. It will handle requests to the "/pins" path and
// related subpaths, so the user does not need to strip the prefix in
// order to use it.
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
	ctx := req.Context()
	token, ok := tokenFromRequest(req)
	if !ok {
		respondError(
			rw,
			http.StatusUnauthorized,
			"no bearer token provided",
		)
		return
	}
	ctx = withToken(ctx, token)

	q := req.URL.Query()
	query := defaultPinQuery()

	query.CID = strings.SplitN(q.Get("cid"), ",", 11)
	if len(query.CID) > 10 {
		respondError(
			rw,
			http.StatusBadRequest,
			fmt.Sprintf("too many CIDs: %v", len(query.CID)),
		)
		return
	}

	query.Name = q.Get("name")

	match := TextMatchingStrategy(q.Get("match"))
	if match != "" {
		if !match.valid() {
			respondError(
				rw,
				http.StatusBadRequest,
				fmt.Sprintf("invalid matching strategy: %q", match),
			)
			return
		}
		query.Match = match
	}

	status := strings.SplitN(q.Get("status"), ",", 5)
	if (len(status) == 0) || (len(status) > 4) {
		respondError(
			rw,
			http.StatusBadRequest,
			"status list must be non-empty and have at most 4 elements",
		)
		return
	}
	for _, v := range status {
		query.Status = append(query.Status, RequestStatus(v))
	}

	before := q.Get("before")
	if before != "" {
		var err error
		query.Before, err = time.Parse(time.RFC3339, before)
		if err != nil {
			respondError(
				rw,
				http.StatusBadRequest,
				fmt.Sprintf("invalid before %q: %v", before, err),
			)
			return
		}
	}

	after := q.Get("after")
	if after != "" {
		var err error
		query.After, err = time.Parse(time.RFC3339, after)
		if err != nil {
			respondError(
				rw,
				http.StatusBadRequest,
				fmt.Sprintf("invalid after %q: %v", after, err),
			)
			return
		}
	}

	limit := q.Get("limit")
	if limit != "" {
		plimit, err := strconv.ParseInt(limit, 10, 0)
		if err != nil {
			respondError(
				rw,
				http.StatusBadRequest,
				fmt.Sprintf("invalid limit %q: %v", limit, err),
			)
			return
		}
		query.Limit = int(plimit)
	}

	meta := q.Get("meta")
	if meta != "" {
		err := json.Unmarshal([]byte(meta), &query.Meta)
		if err != nil {
			respondError(
				rw,
				http.StatusBadRequest,
				fmt.Sprintf("invalid meta %q: %v", meta, err),
			)
			return
		}
	}

	pins, err := h.h.Pins(ctx, query)
	if err != nil {
		respondError(
			rw,
			http.StatusInternalServerError,
			"",
		)
		return
	}

	err = json.NewEncoder(rw).Encode(struct {
		Count   int         `json:"count"`
		Results []PinStatus `json:"results"`
	}{
		Count:   len(pins),
		Results: pins,
	})
	if err != nil {
		respondError(
			rw,
			http.StatusInternalServerError,
			"",
		)
		return
	}
}

func (h handler) postPins(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	token, ok := tokenFromRequest(req)
	if !ok {
		respondError(
			rw,
			http.StatusUnauthorized,
			"no bearer token provided",
		)
		return
	}
	ctx = withToken(ctx, token)

	q := req.URL.Query()
	var pin Pin

	pin.CID = q.Get("cid")
	if pin.CID == "" {
		respondError(
			rw,
			http.StatusBadRequest,
			"CID is required",
		)
		return
	}

	pin.Name = q.Get("name")
	if len(pin.Name) > 255 {
		respondError(
			rw,
			http.StatusBadRequest,
			fmt.Sprintf("name length of %v is longer than the maximum of 255", len(pin.Name)),
		)
		return
	}

	pin.Origins = strings.SplitN(q.Get("origins"), ",", 21)
	if len(pin.Origins) > 20 {
		respondError(
			rw,
			http.StatusBadRequest,
			"maximum of 20 origins allowed",
		)
		return
	}

	meta := q.Get("meta")
	if meta != "" {
		err := json.Unmarshal([]byte(meta), &pin.Meta)
		if err != nil {
			respondError(
				rw,
				http.StatusBadRequest,
				fmt.Sprintf("invalid meta %q: %v", meta, err),
			)
			return
		}
	}

	status, err := h.h.AddPin(ctx, pin)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, "")
		return
	}

	err = json.NewEncoder(rw).Encode(status)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, "")
		return
	}
}

func (h handler) getPinByID(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	token, ok := tokenFromRequest(req)
	if !ok {
		respondError(
			rw,
			http.StatusUnauthorized,
			"no bearer token provided",
		)
		return
	}
	ctx = withToken(ctx, token)

	vars := mux.Vars(req)
	id := vars["requestID"]
	if id == "" {
		respondError(
			rw,
			http.StatusBadRequest,
			"request ID is required",
		)
		return
	}

	status, err := h.h.GetPin(ctx, id)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, "")
		return
	}

	err = json.NewEncoder(rw).Encode(status)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, "")
		return
	}
}

func (h handler) postPinByID(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

func (h handler) deletePinByID(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

type errorResponse struct {
	Error errorResponseError `json:"error"`
}

type errorResponseError struct {
	Reason  string `json:"reason"`
	Details string `json:"details,omitempty"`
}

func respondError(rw http.ResponseWriter, status int, err string) {
	var buf strings.Builder
	json.NewEncoder(&buf).Encode(errorResponse{
		Error: errorResponseError{
			Reason:  reasonFromStatus(status),
			Details: err,
		},
	})
	http.Error(rw, buf.String(), status)
}

func reasonFromStatus(status int) string {
	// TODO: Handle more statuses?
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"

	default:
		return "INTERNAL_ERROR"
	}
}
