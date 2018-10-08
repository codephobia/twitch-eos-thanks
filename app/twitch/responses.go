package twitch

import database "github.com/codephobia/twitch-eos-thanks/app/database"

// follower list response
type FollowersResp struct {
	Data []struct {
		FollowerID string `json:"followerID"`
		Timestamp  string `json:"timestamp"`
	} `json:"data"`
}

// subscriber list response
type SubscribersResp struct {
	Data []*database.Subscriber `json:"data"`
}

// bit list response
type BitsResp struct {
	Data []*database.Bit `json:"data"`
}

// twitch user response
type UserResp struct {
	Data []*TwitchUser `json:"data"`
}

// twitch streams response
type StreamsResp struct {
	Stream struct {
		CreatedAt string `json:"created_at"`
	} `json:"stream"`
}
