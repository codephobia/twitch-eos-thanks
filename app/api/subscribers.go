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
	// subscribers to return
	subscribers := make([]database.Subscriber, 0)

	// add headers to response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// db subscribers
	dbSubscribers := make([][]byte, 0)

	// if limiting subscribers to current stream
	if api.config.ClientShowCurrentStream {
		// get current stream subscribers from db
		// TODO: fix followed_at
		err, f := api.database.GetAllSince(twitch.TWITCH_SUBSCRIBER_DB_BUCKET, api.twitch.StreamStartTime, "followed_at")
		if err != nil {
			api.handleError(w, 500, err)
			return
		}

		// set subscribers
		dbSubscribers = f
	} else {
		// load all subscribers from db
		err, f := api.database.GetAll(twitch.TWITCH_SUBSCRIBER_DB_BUCKET)
		if err != nil {
			api.handleError(w, 500, err)
			return
		}

		// set subscribers
		dbSubscribers = f
	}

	// unmarshal db subscribers
	for _, dbSubscriber := range dbSubscribers {
		var subscriber database.Subscriber
		if err := json.Unmarshal(dbSubscriber, &subscriber); err != nil {
			api.handleError(w, 500, err)
			return
		}

		// append to subscribers returned
		subscribers = append(subscribers, subscriber)
	}

	// encode the subscribers
	enc := json.NewEncoder(w)
	enc.Encode(subscribers)
}
