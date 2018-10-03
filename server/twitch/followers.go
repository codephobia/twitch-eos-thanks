package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/codephobia/twitch-eos-thanks/server/database"
)

// Follower is a Twitch follower.
type Follower struct {
	FollowerID string `json:"from_id"`
	FollowedAt string `json:"followed_at"`
	UserData   *TwitchUser
}

// FollowerResp is a Twitch response of followers
type FollowerResp struct {
	Data []struct {
		FollowedAt string `json:"followed_at"`
		FromID     string `json:"from_id"`
		ToID       string `json:"to_id"`
	} `json:"data"`

	Total      int `json:"total"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

// get the current followers for the channel and
func (t *Twitch) getFollowers() error {
	// build query url
	urlSuffix := strings.Join([]string{TWITCH_HELIX_FOLLOWERS_URL, t.config.TwitchChannelID, "&first=100"}, "")

	// track follower count for loop check
	followerCount := 0

	// toggle loop
	loop := true

	// twitch api pagination cursor
	cursor := ""

	// loop through pages of twitch followers
	for loop {
		// url suffix with possible pagination cursor
		urlPaginate := urlSuffix

		// check for cursor and add to url
		if len(cursor) > 0 {
			urlPaginate += "&after=" + cursor
		}

		// get followers from twitch
		body, err := t.getTwitchResponse(TwitchHelix, urlPaginate)
		if err != nil {
			return err
		}

		// decode data into followers
		followerResp := &FollowerResp{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(followerResp); err != nil {
			return fmt.Errorf("body decode: %s", err)
		}

		// make sure we got data
		if len(followerResp.Data) == 0 {
			return fmt.Errorf("no data: %s", string(body))
		}

		// save follower data to datbase
		for _, follower := range followerResp.Data {
			// parse follow time
			timestamp, err := time.Parse(time.RFC3339, follower.FollowedAt)
			if err != nil {
				log.Printf("unable to parse follower timestamp: %s", err)
				continue
			}

			// make new follower
			f := &database.Follower{
				ChannelID:  follower.ToID,
				FollowerID: follower.FromID,
				Timestamp:  timestamp,
			}

			// add new follower to database
			err = t.database.AddFollower(f)
			if err != nil {
				log.Printf("unable to add follower [%s] to channel [%s]: %s", f.FollowerID, f.ChannelID, err)
			}
		}

		// update cursor
		cursor = followerResp.Pagination.Cursor

		// update follower count
		followerCount += len(followerResp.Data)

		// check for loop end
		if followerCount >= followerResp.Total {
			loop = false
		}

		// sleep between api calls
		time.Sleep(TWITCH_API_DELAY)
	}

	return nil
}
