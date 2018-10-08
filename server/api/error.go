package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResp is an error response.
type ErrorResp struct {
	Err string `json:"error"`
}

// handle an error response
func (api *API) handleError(w http.ResponseWriter, status int, err error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.Encode(&ErrorResp{
		Err: err.Error(),
	})
}
