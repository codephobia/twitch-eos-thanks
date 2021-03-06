package api

import (
    "fmt"
    "encoding/json"
    "net/http"
)

type ApiSettings struct {
    ClientTimeTotal         int  `json:"clientTimeTotal"`
    ClientTimePer           int  `json:"clientTimePer"`
    ClientShowFollowers     bool `json:"clientShowFollowers"`
    ClientShowSubscribers   bool `json:"clientShowSubscribers"`
    ClientShowCurrentStream bool `json:"clientShowCurrentStream"`
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
        ClientTimeTotal:         api.config.ClientTimeTotal,
        ClientTimePer:           api.config.ClientTimePer,
        ClientShowFollowers:     api.config.ClientShowFollowers,
        ClientShowSubscribers:   api.config.ClientShowSubscribers,
        ClientShowCurrentStream: api.config.ClientShowCurrentStream,
    }
    
    // encode the settings
    enc := json.NewEncoder(w)
    enc.Encode(apiSettings)
}
