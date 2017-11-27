package api

import (
    "fmt"
    "encoding/json"
    "net/http"
)

type ApiSettings struct {
    ClientTotalTime int `json:"clientTotalTime"`
}

// handleSettings
func (api *Api) handleSettings() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
            case "GET":
                api.handleSettingsGet(w, r)
            default:
                api.handleError(w, 400, fmt.Errorf("method not allowed"))
        }
    })
}

// handleSettingsGet
func (api *Api) handleSettingsGet(w http.ResponseWriter, r *http.Request) {
    // add headers to response
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // generate our api safe settings
    apiSettings := &ApiSettings{
        ClientTotalTime: api.config.ClientTotalTime,
    }
    
    // encode the settings
    enc := json.NewEncoder(w)
    enc.Encode(apiSettings)
}
