package twitch

import (
	"fmt"
	"log"
	"time"

	config "github.com/codephobia/twitch-eos-thanks/app/config"
	database "github.com/codephobia/twitch-eos-thanks/app/database"
	util "github.com/codephobia/twitch-eos-thanks/app/util"
)

var (
	TWITCH_API_DELAY          time.Duration = 500 * time.Millisecond
	TWITCH_API_CRON_DURATION  time.Duration = 5 * time.Minute
	TWITCH_API_FOLLOWER_LIMIT int           = 100
	TWITCH_API_USER_LIMIT     int           = 100

	TWITCH_HELIX_USERS_URL string = "/users?"

	TWITCH_DB_BUCKET          []string = []string{"twitch"}
	TWITCH_FOLLOWER_DB_BUCKET []string = append(TWITCH_DB_BUCKET, "followers")
)

// twitch
type Twitch struct {
	config   *config.Config
	database *database.Database
	timer    *util.Timer

	Followers       []*Follower
	StreamStartTime time.Time
}

// create twitch
func NewTwitch(c *config.Config, db *database.Database) (*Twitch, error) {
	// init the bucket
	if err := db.InitBucket(TWITCH_FOLLOWER_DB_BUCKET); err != nil {
		return nil, fmt.Errorf("init twitch bucket: ", err)
	}

	// return new twitch struct
	return &Twitch{
		config:   c,
		database: db,
	}, nil
}

// get data from helix
func (t *Twitch) Get() error {
	// get stream start time
	if streamTime, err := t.getStreamStart(); err != nil {
		return err
	} else {
		t.StreamStartTime = streamTime
	}

	// get follower ids
	if err := t.getFollowers(); err != nil {
		return err
	}

	// get user data from follower ids
	if err := t.getFollowerUserData(); err != nil {
		return err
	}

	// save the followers to the database
	if err := t.saveFollowers(); err != nil {
		return err
	}

	// run timer to poll twitch api
	t.startTimer()

	return nil
}

// start a timer for twitch api polling
func (t *Twitch) startTimer() {
	t.timer = util.NewTimer(TWITCH_API_CRON_DURATION, false, t.cron)
}

// cron function run by timer
func (t *Twitch) cron() {
	// get twitch followers
	err := t.Get()
	if err != nil {
		log.Printf("twitch cron: %s", err)
	}
}
