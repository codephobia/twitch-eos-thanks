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
	TWITCH_API_DELAY            time.Duration = 500 * time.Millisecond
	TWITCH_API_CRON_DURATION    time.Duration = 5 * time.Minute
	TWITCH_API_FOLLOWER_LIMIT   int           = 100
	TWITCH_API_SUBSCRIBER_LIMIT int           = 100
	TWITCH_API_BITS_LIMIT       int           = 100
	TWITCH_API_USER_LIMIT       int           = 100

	TWITCH_HELIX_USERS_URL string = "/users?"

	TWITCH_DB_BUCKET            []string = []string{"twitch"}
	TWITCH_FOLLOWER_DB_BUCKET   []string = append(TWITCH_DB_BUCKET, "followers")
	TWITCH_SUBSCRIBER_DB_BUCKET []string = append(TWITCH_DB_BUCKET, "subscribers")
	TWITCH_BIT_DB_BUCKET        []string = append(TWITCH_DB_BUCKET, "bits")
)

// twitch
type Twitch struct {
	config   *config.Config
	database *database.Database
	timer    *util.Timer

	Followers       []*Follower
	Subscribers     []*database.Subscriber
	Bits            []*database.Bit
	StreamStartTime time.Time
}

// create twitch
func NewTwitch(c *config.Config, db *database.Database) (*Twitch, error) {
	// init the followers bucket
	if err := db.InitBucket(TWITCH_FOLLOWER_DB_BUCKET); err != nil {
		return nil, fmt.Errorf("init twitch followers bucket: %s", err)
	}

	// init the subscribers bucket
	if err := db.InitBucket(TWITCH_SUBSCRIBER_DB_BUCKET); err != nil {
		return nil, fmt.Errorf("init twitch subscribers bucket: %s", err)
	}

	// init the bits bucket
	if err := db.InitBucket(TWITCH_BIT_DB_BUCKET); err != nil {
		return nil, fmt.Errorf("init twitch bits bucket: %s", err)
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

	// get subscribers
	if err := t.getSubscribers(); err != nil {
		return err
	}

	// save the subscribers to the database
	if err := t.saveSubscribers(); err != nil {
		return err
	}

	// get bits
	if err := t.getBits(); err != nil {
		return err
	}

	// save the bits to the database
	if err := t.saveBits(); err != nil {
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
