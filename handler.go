package sips

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	errNoToken            = errors.New("no bearer token provided")
	errInvalidStatusQuery = errors.New("status list must be non-empty and have at most 4 elements")
	errNoRequestID        = errors.New("request ID is required")
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
//
// Every method is called after the authentication token is pulled
// from HTTP headers, so it can be assumed that a token is included in
// the provided context. It should not, however, be assumed that the
// token is valid.
//
// Errors returned by a PinHandler's methods are returned to the
// client verbatim, so implementations should be careful not to
// include data that shouldn't be shown to clients in them. If an
// error implements StatusError, the HTTP status code returned will be
// whatever the error's Status method returns.
type PinHandler interface {
	// Pins returns a list of pinning request statuses based on the
	// given query.
	//
	// BUG: Doesn't properly allow a differenation between number of
	// results returned and total number of results, thus breaking
	// paging.
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
	r := httprouter.New()
	r.HandleOPTIONS = true
	r.GlobalOPTIONS = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Access-Control-Request-Method") != "" {
			wh := rw.Header()
			wh.Set("Access-Control-Allow-Methods", wh.Get("Allow"))
			wh.Set("Access-Control-Allow-Origin", "*")
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	handler := handler{h: h}
	r.GET("/pins", handler.getPins)
	r.POST("/pins", handler.postPins)
	r.GET("/pins/:requestID", handler.getPinByID)
	r.POST("/pins/:requestID", handler.postPinByID)
	r.DELETE("/pins/:requestID", handler.deletePinByID)

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		token, ok := tokenFromRequest(req)
		if !ok {
			respondError(
				rw,
				http.StatusUnauthorized,
				errNoToken,
			)
			return
		}
		req = req.WithContext(withToken(req.Context(), token))

		r.ServeHTTP(rw, req)
	})
}

func (h handler) getPins(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := req.Context()

	q := req.URL.Query()
	query := defaultPinQuery()

	query.CID = strings.SplitN(q.Get("cid"), ",", 11)
	switch {
	case len(query.CID) == 1:
		if query.CID[0] == "" {
			query.CID = nil
		}

	case len(query.CID) > 10:
		respondError(
			rw,
			http.StatusBadRequest,
			fmt.Errorf("too many CIDs: %v", len(query.CID)),
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
				fmt.Errorf("invalid matching strategy: %q", match),
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
			errInvalidStatusQuery,
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
				fmt.Errorf("invalid before %q: %w", before, err),
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
				fmt.Errorf("invalid after %q: %w", after, err),
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
				fmt.Errorf("invalid limit %q: %w", limit, err),
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
				fmt.Errorf("invalid meta %q: %w", meta, err),
			)
			return
		}
	}

	pins, err := h.h.Pins(ctx, query)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
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
		respondError(rw, http.StatusInternalServerError, err)
		return
	}
}

func (h handler) postPins(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}

	var pin Pin
	err = json.Unmarshal(body, &pin)
	if err != nil {
		respondError(
			rw,
			http.StatusBadRequest,
			fmt.Errorf("failed to parse body: %w", err),
		)
		return
	}

	status, err := h.h.AddPin(ctx, pin)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}

	err = json.NewEncoder(rw).Encode(status)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}
}

func (h handler) getPinByID(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := req.Context()

	id := params[0].Value
	if id == "" {
		respondError(
			rw,
			http.StatusBadRequest,
			errNoRequestID,
		)
		return
	}

	status, err := h.h.GetPin(ctx, id)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}

	err = json.NewEncoder(rw).Encode(status)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}
}

func (h handler) postPinByID(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := req.Context()

	id := params[0].Value
	if id == "" {
		respondError(
			rw,
			http.StatusBadRequest,
			errNoRequestID,
		)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}

	var pin Pin
	err = json.Unmarshal(body, &pin)
	if err != nil {
		respondError(
			rw,
			http.StatusBadRequest,
			fmt.Errorf("failed to parse body: %w", err),
		)
		return
	}

	status, err := h.h.UpdatePin(ctx, id, pin)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}

	err = json.NewEncoder(rw).Encode(status)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}
}

func (h handler) deletePinByID(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := req.Context()

	id := params[0].Value
	if id == "" {
		respondError(
			rw,
			http.StatusBadRequest,
			errNoRequestID,
		)
		return
	}

	err := h.h.DeletePin(ctx, id)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, err)
		return
	}

	// Yields no response body if successful.
}

type errorResponse struct {
	Error errorResponseError `json:"error"`
}

type errorResponseError struct {
	Reason  string `json:"reason"`
	Details string `json:"details,omitempty"`
}

func respondError(rw http.ResponseWriter, status int, err error) {
	var statusError StatusError
	if errors.As(err, &statusError) {
		status = statusError.Status()
	}

	rw.WriteHeader(status)

	json.NewEncoder(rw).Encode(errorResponse{
		Error: errorResponseError{
			Reason:  reasonFromStatus(status),
			Details: err.Error(),
		},
	})
}

func reasonFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"

	case http.StatusUnauthorized:
		return "UNAUTHORIZED"

	case http.StatusNotFound:
		return "NOT_FOUND"

	case http.StatusConflict:
		return "INSUFFICIENT_FUNDS"

	default:
		return "INTERNAL_SERVER_ERROR"
	}
}

// StatusError is implemented by errors returned by PinHandler
// implementations that want to send custom status codes to the
// client.
//
// Several status codes have special handling. These include
//   - 400 Bad Request
//   - 401 Unauthorized
//   - 404 Not Found
//   - 409 Conflict
//
// These status codes will produce special error messages for the
// client. All other status codes will produce the same error message
// as a 500 Internal Server Error code does.
type StatusError interface {
	Status() int
}
