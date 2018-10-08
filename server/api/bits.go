package api

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

// handleBits
func (api *API) handleBits() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			api.handleBitsGet(w, r)
		default:
			api.handleError(w, 400, fmt.Errorf("method not allowed"))
		}
	})
}

// handleBitsGet
func (api *API) handleBitsGet(w http.ResponseWriter, r *http.Request) {
	var (
		limitDefault  = 20
		limitMax      = 100
		offsetDefault = 0
	)

	// get query vars
	v := r.URL.Query()

	// get vars
	// TODO: error check this
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
	if offset <= offsetDefault {
		offset = offsetDefault
	}

	// get bits
	bits, err := api.database.GetBits(channelID, latest, limit, offset)
	if err != nil {
		log.Printf("[ERROR] get bits: %s", err)
	}

	api.handleSuccess(w, bits)
}
