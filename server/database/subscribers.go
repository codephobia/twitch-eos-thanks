package database

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Subscriber is a twitch follower.
type Subscriber struct {
	ID           bson.ObjectId `bson:"_id,omitempty" json:"ID,omitempty"`
	ChannelID    string        `bson:"channel_id,omitempty" json:"channelID,omitempty"`
	SubscriberID string        `bson:"subscriber_id,omitempty" json:"subscriberID,omitempty"`
	Timestamp    time.Time     `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
}

// AddSubscriber adds a subscriber to the database.
func (db *Database) AddSubscriber(s *Subscriber) error {
	// check if subscriber already exists
	subscribed, err := db.hasSubscriber(s.ChannelID, s.SubscriberID)
	if err != nil {
		return err
	}

	// skip adding subscriber to database if they are already subscribed
	if subscribed {
		return fmt.Errorf("found duplicate subscriber [%s] for channel [%s]", s.SubscriberID, s.ChannelID)
	}

	// insert new subscriber
	return db.subscribers.Insert(s)
}

// check for subscriber
func (db *Database) hasSubscriber(channelID string, subscriberID string) (bool, error) {
	// build query
	query := db.subscribers.Find(bson.M{
		"channel_id":    channelID,
		"subscriber_id": subscriberID,
	})

	// get count
	count, err := query.Count()
	if err != nil {
		return false, err
	}

	// return that we found a subscriber
	if count > 0 {
		return true, nil
	}

	// return that we didn't find a subscriber
	return false, nil
}

// RemoveSubscriber removes a subscriber from the database.
func (db *Database) RemoveSubscriber(s *Subscriber) error {
	// remove the subscriber from the database
	return db.subscribers.Remove(bson.M{
		"channel_id":    s.ChannelID,
		"subscriber_id": s.SubscriberID,
	})
}

// GetSubscribers returns a slice of subscribers.
func (db *Database) GetSubscribers(channelID string, latest int64, limit int, offset int) ([]*Subscriber, error) {
	subscribers := make([]*Subscriber, 0)

	// convert latest to time
	ltUnix := latest / (int64(time.Millisecond) * int64(time.Nanosecond) * 1000)
	ltNano := latest % (int64(time.Millisecond) * int64(time.Nanosecond) * 1000)
	lt := time.Unix(ltUnix, ltNano)

	// build query
	query := db.subscribers.Find(bson.M{
		"channel_id": channelID,
		"timestamp": bson.M{
			"$gt": lt,
		},
	})

	// add filters
	query.Limit(limit).Skip(offset).Sort("-timestamp").Select(bson.M{
		"_id":           0,
		"subscriber_id": 1,
		"timestamp":     1,
	})

	// get subscribers
	err := query.All(&subscribers)
	if err != nil {
		return subscribers, fmt.Errorf("unable to get subscribers: %s", err)
	}

	return subscribers, nil
}
