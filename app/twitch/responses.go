package twitch

// follower list response
type FollowersResp struct {
    Data []struct {
        FollowerID string `json:"followerID"`
        Timestamp  string `json:"timestamp"`        
    } `json:"data"`
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