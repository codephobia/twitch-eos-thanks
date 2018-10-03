package twitch

import (
	"time"

	"github.com/codephobia/twitch-eos-thanks/server/config"
	"github.com/codephobia/twitch-eos-thanks/server/database"
)

var (
	TWITCH_API_DELAY          time.Duration = 500 * time.Millisecond
	TWITCH_API_FOLLOWER_LIMIT int           = 100
	TWITCH_API_USER_LIMIT     int           = 100

	TWITCH_HELIX_USERS_URL     string = "/users?"
	TWITCH_HELIX_FOLLOWERS_URL string = "/users/follows?to_id="
)

// twitch
type Twitch struct {
	config   *config.Config
	database *database.Database
}

// create twitch
func NewTwitch(c *config.Config, db *database.Database) *Twitch {
	return &Twitch{
		config:   c,
		database: db,
	}
}

// Init initializes the twitch channels, getting followers if need be
func (t *Twitch) Init() error {
	// check if we have followers already
	hasFollowers, err := t.database.HasFollowers(t.config.TwitchChannelID)
	if err != nil {
		return err
	}

	// if we already have followers, skip getting followers
	if hasFollowers {
		return nil
	}

	// get followers from twitch
	return t.getFollowers()
}
