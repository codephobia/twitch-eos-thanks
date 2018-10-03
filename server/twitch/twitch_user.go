package twitch

type TwitchUser struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	ProfileImageUrl string `json:"profile_image_url"`
}
