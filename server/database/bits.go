package database

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Bit is a bit pub sub message from twitch.
type Bit struct {
	ID               bson.ObjectId     `bson:"_id,omitempty" json:"ID,omitempty"`
	UserName         string            `bson:"user_name" json:"user_name"`
	ChannelName      string            `bson:"channel_name" json:"channel_name"`
	UserID           string            `bson:"user_id" json:"user_id"`
	ChannelID        string            `bson:"channel_id" json:"channel_id"`
	Time             time.Time         `bson:"timestamp" json:"timestamp"`
	ChatMessage      string            `bson:"chat_message" json:"chat_message"`
	BitsUsed         int               `bson:"bits_used" json:"bits_used"`
	TotalBitsUsed    int               `bson:"total_bits_used" json:"total_bits_used"`
	Context          string            `bson:"context" json:"context"`
	BadgeEntitlement *BadgeEntitlement `bson:"badge_entitlement" json:"badge_entitlement"`
}

// BadgeEntitlement contains meta data for the user badge on a Bit.
type BadgeEntitlement struct {
	NewVersion      int `bson:"new_version" json:"new_version"`
	PreviousVersion int `bson:"previous_version" json:"previous_version"`
}

// AddBit adds a bit event to the database.
func (db *Database) AddBit(b *Bit) error {
	// insert new bit event
	return db.bits.Insert(b)
}

// GetBits returns a slice of bit events.
func (db *Database) GetBits(channelID string, latest int64, limit int, offset int) ([]*Bit, error) {
	bits := make([]*Bit, 0)

	// convert latest to time
	ltUnix := latest / (int64(time.Millisecond) * int64(time.Nanosecond) * 1000)
	ltNano := latest % (int64(time.Millisecond) * int64(time.Nanosecond) * 1000)
	lt := time.Unix(ltUnix, ltNano)

	// build query
	query := db.bits.Find(bson.M{
		"channel_id": channelID,
		"timestamp": bson.M{
			"$gt": lt,
		},
	})

	// add filters
	query.Limit(limit).Skip(offset).Sort("-timestamp")

	// get bit events
	err := query.All(&bits)
	if err != nil {
		return bits, fmt.Errorf("unable to get bits: %s", err)
	}

	return bits, nil
}
