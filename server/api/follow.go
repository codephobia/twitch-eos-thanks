package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	database "github.com/codephobia/twitch-eos-thanks/server/database"
)

// Follow is a twitch follow.
type Follow struct {
	Data []struct {
		ToID      string `json:"to_id"`
		FromID    string `json:"from_id"`
		Timestamp string `json:"followed_at"`
	} `json:"data"`
}

// handleFollow
func (api *API) handleFollow() http.Handler {
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
func (api *API) handleFollowGet(w http.ResponseWriter, r *http.Request) {
	// get query vars
	v := r.URL.Query()

	// TODO: add secret validation
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

	// close body
	r.Body.Close()
}

// handleFollowPost
func (api *API) handleFollowPost(w http.ResponseWriter, r *http.Request) {
	// add headers to response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// decode the notification
	var f Follow
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&f)
	if err != nil {
		log.Printf("[ERROR] follow: unable to decode notification: %s", err)
		return
	}

	// loop through all follows on payload
	for _, newFollow := range f.Data {
		// convert time from string
		t, err := time.Parse(time.RFC3339, newFollow.Timestamp)
		if err != nil {
			log.Printf("[ERROR] follow: unable to parse timetamp: %s", err)
			return
		}

		// build follower for db
		follower := &database.Follower{
			ChannelID:  newFollow.ToID,
			FollowerID: newFollow.FromID,
			Timestamp:  t,
		}

		if err := api.database.AddFollower(follower); err != nil {
			log.Printf("[ERROR] unable to add follower: %s", err)
		}
	}

	// close body
	r.Body.Close()
}
