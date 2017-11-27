package api

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type ApiCheck struct {
    Status string `json:"status"`
}

// handleCheck
func (api *Api) handleCheck() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
            case "GET":
                api.handleCheckGet(w, r)
            default:
                api.handleError(w, 400, fmt.Errorf("method not allowed"))
        }
    })
}

// handleCheckGet
func (api *Api) handleCheckGet(w http.ResponseWriter, r *http.Request) {
    // add headers to response
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // return status ok
    apiCheck := &ApiCheck{
        Status: "ok",
    }
    
    // encode the status
    enc := json.NewEncoder(w)
    enc.Encode(apiCheck)
}
