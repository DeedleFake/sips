package sips

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// A PinsApiController binds http requests to an api service and writes the service results to the http response
type PinsApiController struct {
	service PinsApiServicer
}

// NewPinsApiController creates a default api controller
func NewPinsApiController(s PinsApiServicer) Router {
	return &PinsApiController{service: s}
}

// Routes returns all of the api route for the PinsApiController
func (c *PinsApiController) Routes() Routes {
	return Routes{
		{
			"PinsGet",
			strings.ToUpper("Get"),
			"/pins",
			c.PinsGet,
		},
		{
			"PinsPost",
			strings.ToUpper("Post"),
			"/pins",
			c.PinsPost,
		},
		{
			"PinsRequestidDelete",
			strings.ToUpper("Delete"),
			"/pins/{requestid}",
			c.PinsRequestidDelete,
		},
		{
			"PinsRequestidGet",
			strings.ToUpper("Get"),
			"/pins/{requestid}",
			c.PinsRequestidGet,
		},
		{
			"PinsRequestidPost",
			strings.ToUpper("Post"),
			"/pins/{requestid}",
			c.PinsRequestidPost,
		},
	}
}

// PinsGet - List pin objects
func (c *PinsApiController) PinsGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	cid := strings.Split(query.Get("cid"), ",")
	name := query.Get("name")
	match := query.Get("match")
	status := strings.Split(query.Get("status"), ",")
	before := query.Get("before")
	after := query.Get("after")
	limit, err := parseInt32Parameter(query.Get("limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	meta := query.Get("meta")
	result, err := c.service.PinsGet(r.Context(), cid, name, match, status, before, after, limit, meta)
	//If an error occured, encode the error with the status code
	if err != nil {
		EncodeJSONResponse(err.Error(), &result.Code, w)
		return
	}
	//If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PinsPost - Add pin object
func (c *PinsApiController) PinsPost(w http.ResponseWriter, r *http.Request) {
	pin := &Pin{}
	if err := json.NewDecoder(r.Body).Decode(&pin); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := c.service.PinsPost(r.Context(), *pin)
	//If an error occured, encode the error with the status code
	if err != nil {
		EncodeJSONResponse(err.Error(), &result.Code, w)
		return
	}
	//If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PinsRequestidDelete - Remove pin object
func (c *PinsApiController) PinsRequestidDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	requestid := params["requestid"]
	result, err := c.service.PinsRequestidDelete(r.Context(), requestid)
	//If an error occured, encode the error with the status code
	if err != nil {
		EncodeJSONResponse(err.Error(), &result.Code, w)
		return
	}
	//If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PinsRequestidGet - Get pin object
func (c *PinsApiController) PinsRequestidGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	requestid := params["requestid"]
	result, err := c.service.PinsRequestidGet(r.Context(), requestid)
	//If an error occured, encode the error with the status code
	if err != nil {
		EncodeJSONResponse(err.Error(), &result.Code, w)
		return
	}
	//If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PinsRequestidPost - Replace pin object
func (c *PinsApiController) PinsRequestidPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	requestid := params["requestid"]
	pin := &Pin{}
	if err := json.NewDecoder(r.Body).Decode(&pin); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := c.service.PinsRequestidPost(r.Context(), requestid, *pin)
	//If an error occured, encode the error with the status code
	if err != nil {
		EncodeJSONResponse(err.Error(), &result.Code, w)
		return
	}
	//If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
