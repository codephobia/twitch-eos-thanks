package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	database "github.com/codephobia/twitch-eos-thanks/app/database"
	twitch "github.com/codephobia/twitch-eos-thanks/app/twitch"
)

// handleSubscribers
func (api *Api) handleSubscribers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			api.handleSubscribersGet(w, r)
		default:
			api.handleError(w, 400, fmt.Errorf("method not allowed"))
		}
	})
}

// handleSubscribersGet
func (api *Api) handleSubscribersGet(w http.ResponseWriter, r *http.Request) {
	// Subscribers to return
	Subscribers := make([]database.Subscriber, 0)

	// add headers to response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// db Subscribers
	dbSubscribers := make([][]byte, 0)

	// if limiting Subscribers to current stream
	if api.config.ClientShowCurrentStream {
		// get current stream Subscribers from db
		err, f := api.database.GetAllSince(twitch.TWITCH_SUBSCRIBER_DB_BUCKET, api.twitch.StreamStartTime, "followed_at")
		if err != nil {
			api.handleError(w, 500, err)
			return
		}

		// set Subscribers
		dbSubscribers = f
	} else {
		// load all Subscribers from db
		err, f := api.database.GetAll(twitch.TWITCH_SUBSCRIBER_DB_BUCKET)
		if err != nil {
			api.handleError(w, 500, err)
			return
		}

		// set Subscribers
		dbSubscribers = f
	}

	// unmarshal db Subscribers
	for _, dbSubscriber := range dbSubscribers {
		var follower database.Subscriber
		if err := json.Unmarshal(dbSubscriber, &follower); err != nil {
			api.handleError(w, 500, err)
			return
		}

		// append to Subscribers returned
		Subscribers = append(Subscribers, follower)
	}

	// encode the Subscribers
	enc := json.NewEncoder(w)
	enc.Encode(Subscribers)
}
