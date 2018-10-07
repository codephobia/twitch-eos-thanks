package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResp struct {
	Err string `json:"error"`
}

func (api *Api) handleError(w http.ResponseWriter, status int, err error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.Encode(&ErrorResp{
		Err: err.Error(),
	})
}
