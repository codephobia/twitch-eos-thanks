package database

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Subscriber is a twitch subscriber.
type Subscriber struct {
	ID           bson.ObjectId `json:"ID,omitempty"`
	ChannelID    string        `json:"channelID,omitempty"`
	SubscriberID string        `json:"subscriberID,omitempty"`
	Timestamp    time.Time     `json:"timestamp,omitempty"`

	DisplayName string      `json:"display_name"`
	SubPlan     string      `json:"sub_plan"`
	SubPlanName string      `json:"sub_plan_name"`
	Months      int         `json:"months"`
	Context     string      `json:"context"`
	SubMessage  *SubMessage `json:"sub_message"`
}

// SubMessage is the message sent when a user subscribes.
type SubMessage struct {
	Message string             `json:"message"`
	Emotes  []*SubMessageEmote `json:"emotes"`
}

// SubMessageEmote is meta data for a twitch emote contained
// within a message string.
type SubMessageEmote struct {
	Start int `json:"start"`
	End   int `json:"end"`
	ID    int `json:"id"`
}
