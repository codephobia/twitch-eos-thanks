package database

import (
	"time"
)

// Bit is a bit pub sub message from twitch.
type Bit struct {
	ID               string            `json:"ID,omitempty"`
	UserName         string            `json:"user_name"`
	ChannelName      string            `json:"channel_name"`
	UserID           string            `json:"user_id"`
	ChannelID        string            `json:"channel_id"`
	Time             time.Time         `json:"timestamp"`
	ChatMessage      string            `json:"chat_message"`
	BitsUsed         int               `json:"bits_used"`
	TotalBitsUsed    int               `json:"total_bits_used"`
	Context          string            `json:"context"`
	BadgeEntitlement *BadgeEntitlement `json:"badge_entitlement"`
}

// BadgeEntitlement contains meta data for the user badge on a Bit.
type BadgeEntitlement struct {
	NewVersion      int `json:"new_version"`
	PreviousVersion int `json:"previous_version"`
}
