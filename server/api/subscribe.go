package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	database "github.com/codephobia/twitch-eos-thanks/server/database"
)

// Subscribe is a twitch subscribe.
type Subscribe struct {
	ID    string `json:"id"`
	Topic string `json:"topic"`
	Type  string `json:"type"`
	Data  struct {
		FromID string `json:"from_id"`
		ToID   string `json:"to_id"`
	} `json:"data"`
	Timestamp string `json:"timestamp"`
}

// handleSubscribe
func (api *API) handleSubscribe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			api.handleSubscribeGet(w, r)
		case "POST":
			api.handleSubscribePost(w, r)
		default:
			api.handleError(w, 400, fmt.Errorf("method not allowed"))
		}
	})
}

// handleSubscribeGet
func (api *API) handleSubscribeGet(w http.ResponseWriter, r *http.Request) {
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

// handleSubscribePost
func (api *API) handleSubscribePost(w http.ResponseWriter, r *http.Request) {
	// add headers to response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// decode the notification
	var s Subscribe
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&s)
	if err != nil {
		log.Printf("[ERROR] unable to decode notification: %s", err)
		return
	}

	// convert time from string
	t, err := time.Parse(time.RFC3339, s.Timestamp)
	if err != nil {
		log.Printf("[ERROR] unable to decode notification: %s", err)
		return
	}

	// build subscriber for db
	subscriber := &database.Subscriber{
		ChannelID:    s.Data.ToID,
		SubscriberID: s.Data.FromID,
		Timestamp:    t,
	}

	// determine if we should add or remove the subscriber
	switch s.Type {
	case "create":
		// add the subscriber to database
		if err := api.database.AddSubscriber(subscriber); err != nil {
			log.Printf("[ERROR] unable to add subscriber: %s", err)
		}
	case "delete":
		// remove the subscriber from database
		if err := api.database.RemoveSubscriber(subscriber); err != nil {
			log.Printf("[ERROR] unable to remove subscriber: %s", err)
		}
	default:
		log.Printf("[ERROR] unknown subscribe type: %s", s.Type)
	}
}
