package api

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

// handleFollowers
func (api *API) handleSubscribers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			api.handleSubscribersGet(w, r)
		default:
			api.handleError(w, 400, fmt.Errorf("method not allowed"))
		}
	})
}

// handleFollowersGet
func (api *API) handleSubscribersGet(w http.ResponseWriter, r *http.Request) {
	var (
		limitDefault = 20
		limitMax     = 100
		offsetMax    = 100
	)

	// get query vars
	v := r.URL.Query()

	// get vars
	channelID := v.Get("channelID")
	limit, _ := strconv.Atoi(v.Get("limit"))
	offset, _ := strconv.Atoi(v.Get("offset"))
	latest, _ := strconv.ParseInt(v.Get("latest"), 10, 64)

	// check channel id
	matched, err := regexp.MatchString("[0-9]+", channelID)
	if err != nil || !matched {
		api.handleError(w, 422, fmt.Errorf("invalid channel id"))
		return
	}

	// make sure we have at least default value for limit
	if limit == 0 {
		limit = limitDefault
	}

	// check limit
	if limit > limitMax {
		limit = limitMax
	}

	// check offset
	if offset > offsetMax {
		offset = offsetMax
	}

	// get subscribers
	subscribers, err := api.database.GetSubscribers(channelID, latest, limit, offset)
	if err != nil {
		log.Printf("[ERROR] get subscribers: %s", err)
	}

	api.handleSuccess(w, subscribers)
}
