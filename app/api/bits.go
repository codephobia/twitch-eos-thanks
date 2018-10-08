package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	database "github.com/codephobia/twitch-eos-thanks/app/database"
	twitch "github.com/codephobia/twitch-eos-thanks/app/twitch"
)

// BitResp is a combined bit event.
type BitResp struct {
	DisplayName string `json:"display_name"`
	Bits        int    `json:"bits"`
}

// handleBits
func (api *Api) handleBits() http.Handler {
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
func (api *Api) handleBitsGet(w http.ResponseWriter, r *http.Request) {
	// bits to return
	bits := make([]*BitResp, 0)

	// add headers to response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// db Bits
	dbBits := make([][]byte, 0)

	// if limiting bits to current stream
	if api.config.ClientShowCurrentStream {
		// get current stream bits from db
		// TODO: check followed_at
		err, f := api.database.GetAllSince(twitch.TWITCH_BIT_DB_BUCKET, api.twitch.StreamStartTime, "followed_at")
		if err != nil {
			api.handleError(w, 500, err)
			return
		}

		// set bits
		dbBits = f
	} else {
		// load all bits from db
		err, f := api.database.GetAll(twitch.TWITCH_BIT_DB_BUCKET)
		if err != nil {
			api.handleError(w, 500, err)
			return
		}

		// set bits
		dbBits = f
	}

	// store combined bits
	combinedBits := make(map[string]*BitResp)

	// unmarshal db bits
	for _, dbBit := range dbBits {
		var bit database.Bit
		if err := json.Unmarshal(dbBit, &bit); err != nil {
			api.handleError(w, 500, err)
			return
		}

		// combine bits
		if c, ok := combinedBits[bit.UserID]; !ok {
			combinedBits[bit.UserID] = &BitResp{
				DisplayName: bit.UserName,
				Bits:        bit.BitsUsed,
			}
		} else {
			c.Bits += bit.BitsUsed
		}
	}

	// deconstruct map into array
	for _, cBit := range combinedBits {
		bits = append(bits, cBit)
	}

	// encode the bits
	enc := json.NewEncoder(w)
	enc.Encode(bits)
}
