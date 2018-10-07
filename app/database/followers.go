package database

type Follower struct {
	ID              string `json:"id"`
	FollowedAt      string `json:"followed_at"`
	DisplayName     string `json:"display_name"`
	ProfileImageUrl string `json:"profile_image_url"`
}
