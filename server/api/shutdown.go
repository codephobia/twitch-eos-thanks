package api

import (
    //"context"
    "encoding/json"
    "fmt"
    "net/http"
)

type ApiShutdown struct {
    Status string `json:"status"`
}

// handleShutdown
func (api *Api) handleShutdown() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
            case "GET":
                api.handleShutdownGet(w, r)
            default:
                api.handleError(w, 400, fmt.Errorf("method not allowed"))
        }
    })
}

// handleShutdownGet
func (api *Api) handleShutdownGet(w http.ResponseWriter, r *http.Request) {
    // add headers to response
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // return status ok
    apiShutdown := &ApiShutdown{
        Status: "ok",
    }
    
    // encode the followers
    enc := json.NewEncoder(w)
    enc.Encode(apiShutdown)
    
    //api.server.Shutdown(context.Background())
}
