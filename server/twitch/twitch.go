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

// Twitch ...
type Twitch struct {
	config   *config.Config
	database *database.Database

	pubsub *PUBSUB
}

// NewTwitch returns a new twitch.
func NewTwitch(c *config.Config, db *database.Database) *Twitch {
	twitch := &Twitch{
		config:   c,
		database: db,
	}

	twitch.pubsub = NewPUBSUB(c, db, twitch)

	return twitch
}

// Init initializes the twitch channels, getting followers if need be
func (t *Twitch) Init() error {
	// check if we have followers already
	hasFollowers, err := t.database.HasFollowers(t.config.TwitchChannelID)
	if err != nil {
		return err
	}

	// if we don't have followers, get followers
	if !hasFollowers {
		// get followers from twitch
		if err := t.getFollowers(); err != nil {
			return err
		}
	}

	// init pubsub
	return t.pubsub.Init()
}
