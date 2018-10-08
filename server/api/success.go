package api

import (
	"encoding/json"
	"net/http"
)

// DataResp is an api response.
type DataResp struct {
	Data interface{} `json:"data"`
}

// handle a success response
func (api *API) handleSuccess(w http.ResponseWriter, data interface{}) {
	// add headers to response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// return followers
	enc := json.NewEncoder(w)
	enc.Encode(&DataResp{
		Data: data,
	})
}
