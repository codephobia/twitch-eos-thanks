package twitch

type TwitchVersion uint

const (
	TwitchV5 TwitchVersion = iota
	TwitchHelix
)

func (v TwitchVersion) Url() string {
	return twitchVersionUrl[v]
}

var twitchVersionUrl = map[TwitchVersion]string{
	TwitchV5:    "https://api.twitch.tv/kraken",
	TwitchHelix: "https://api.twitch.tv/helix",
}
