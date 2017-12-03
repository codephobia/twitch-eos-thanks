package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    
    database "github.com/codephobia/twitch-eos-thanks/app/database"
    twitch   "github.com/codephobia/twitch-eos-thanks/app/twitch"
)

// handleFollowers
func (api *Api) handleFollowers() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
            case "GET":
                api.handleFollowersGet(w, r)
            default:
                api.handleError(w, 400, fmt.Errorf("method not allowed"))
        }
    })
}

// handleFollowersGet
func (api *Api) handleFollowersGet(w http.ResponseWriter, r *http.Request) {
    // followers to return
    followers := make([]database.Follower, 0)
    
    // add headers to response
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // db followers
    dbFollowers := make([][]byte, 0)
    
    // if limiting followers to current stream
    if api.config.ClientShowCurrentStream {
        // get current stream followers from db
        err, f := api.database.GetAllSince(twitch.TWITCH_FOLLOWER_DB_BUCKET, api.twitch.StreamStartTime, "followed_at")
        if err != nil {
            api.handleError(w, 500, err)
            return
        }
        
        // set followers
        dbFollowers = f
    } else {
        // load all followers from db
        err, f := api.database.GetAll(twitch.TWITCH_FOLLOWER_DB_BUCKET)
        if err != nil {
            api.handleError(w, 500, err)
            return
        }
        
        // set followers
        dbFollowers = f
    }
    
    // unmarshal db followers
    for _, dbFollower := range dbFollowers {
        var follower database.Follower
        if err := json.Unmarshal(dbFollower, &follower); err != nil {
            api.handleError(w, 500, err)
            return
        }

        // append to followers returned
        followers = append(followers, follower)
    }
    
    // encode the followers
    enc := json.NewEncoder(w)
    enc.Encode(followers)
}
