package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    database "github.com/codephobia/twitch-eos-thanks/server/database"
)

type TwitchHub struct {
    Mode string `json:"mode"`
    Topic string `json:"topic"`
    LeaseSeconds int `json:"lease_seconds,omitempty"`
    Challenge string `json:"challenge"`
    Reason string `json:"reason,omitempty"`
}

type Follow struct {
    ID    string `json:"id"`
    Topic string `json:"topic"`
    Type  string `json:"type"`
    Data struct {
        FromID string `json:"from_id"`
        ToID   string `json:"to_id"`
    } `json:"data"`
    Timestamp string `json:"timestamp"`
}

// handleFollow
func (api *Api) handleFollow() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
            case "GET":
                api.handleFollowGet(w, r)
            case "POST":
                api.handleFollowPost(w, r)
            default:
                api.handleError(w, 400, fmt.Errorf("method not allowed"))
        }
    })
}

// handleFollowGet
func (api *Api) handleFollowGet(w http.ResponseWriter, r *http.Request) {
    // get query vars
    v := r.URL.Query()
    
    // get vars
    hubMode := v.Get("hub.mode")
    //hubTopic := v.Get("hub.topic")
    //hubLeaseSeconds := v.Get("hub.lease_seconds")
    hubChallenge := v.Get("hub.challenge")
    //hubReason := v.Get("hub.reason")
        
    switch hubMode {
        case "subscribe":
            w.WriteHeader(http.StatusAccepted)
            w.Write([]byte(hubChallenge))
        case "denied":
            w.WriteHeader(http.StatusOK)
        default:
            w.WriteHeader(http.StatusOK)
            log.Printf("[ERROR] invalid request mode: %s", hubMode)
    }
}

// handleFollowPost
func (api *Api) handleFollowPost(w http.ResponseWriter, r *http.Request) {
    // add headers to response
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // decode the notification
    var f Follow
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&f)
    if err != nil {
        log.Printf("[ERROR] unable to decode notification: %s", err)
        return
    }
    
    // convert time from string
    t, err := time.Parse(time.RFC3339, f.Timestamp)
    if err != nil {
        log.Printf("[ERROR] unable to decode notification: %s", err)
        return
    }
    
    // build follower for db
    follower := &database.Follower{
        ChannelID:  f.Data.ToID,
        FollowerID: f.Data.FromID,
        Timestamp:  t,
    }
    
    // determine if we should add or remove the follower
    switch f.Type {
        case "create":
            // add the follower to database
            if err := api.database.AddFollower(follower); err != nil {
                log.Printf("[ERROR] unable to add follower: %s", err)
            }
        case "delete":
            // remove the follower from database
            if err := api.database.RemoveFollower(follower); err != nil {
                log.Printf("[ERROR] unable to remove follower: %s", err)
            }
        default:
            log.Printf("[ERROR] unknown follow type: %s", f.Type)
    }
}
