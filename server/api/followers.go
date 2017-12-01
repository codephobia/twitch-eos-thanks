package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "regexp"
    "strconv"
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
    var (
        LIMIT_DEFAULT  int = 20
        LIMIT_MAX      int = 100
        OFFSET_MAX     int = 100
    )
    
    // get query vars
    v := r.URL.Query()
    
    // get vars
    channelID := v.Get("channelID")
    limit, _ := strconv.Atoi(v.Get("limit"))
    offset, _ := strconv.Atoi(v.Get("offset"))
    
    // check channel id
    matched, err := regexp.MatchString("[0-9]+", channelID)
    if (err != nil || !matched) {
        api.handleError(w, 422, fmt.Errorf("invalid channel id"))
        return
    }
    
    // make sure we have at least default value for limit
    if limit == 0 {
        limit = LIMIT_DEFAULT
    }
    
    // check limit
    if (limit > LIMIT_MAX) {
        limit = LIMIT_MAX
    }
    
    // check offset
    if (offset > OFFSET_MAX) {
        offset = OFFSET_MAX
    }
    
    // add headers to response
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // get followers
    followers, err := api.database.GetFollowers(channelID, limit, offset)
    if err != nil {
        log.Printf("[ERROR] get followers: %s", err)
    }
    
    // return followers
    enc := json.NewEncoder(w)
    enc.Encode(followers)
}
