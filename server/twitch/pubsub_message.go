package twitch

import (
	"encoding/json"
)

// PUBSUBMessage is a message when a pub sub event is received.
type PUBSUBMessage struct {
	Type  PUBSUBType         `json:"type"`
	Data  *PUBSUBMessageData `json:"data",omitempty`
	Error string             `json:"error",omitempty`
}

// PUBSUBMessageData is the data on a pub sub message.
type PUBSUBMessageData struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

// NewPUBSUBMessage returns a message from a websocket message.
func NewPUBSUBMessage(message []byte) (*PUBSUBMessage, error) {
	var msg PUBSUBMessage

	// unmarshal message bytes
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, err
	}

	// return new message
	return &msg, nil
}

// ToBytes returns a message as an array of bytes.
func (m *PUBSUBMessage) ToBytes() ([]byte, error) {
	// marshal message in to bytes
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	// return message bytes
	return data, nil
}

// PUBSUBSubscriptionMessage is a message for a subscription
// event on a pub sub message.
type PUBSUBSubscriptionMessage struct {
	UserName    string `json:"user_name"`
	DisplayName string `json:"display_name"`
	ChannelName string `json:"channel_name"`
	UserID      string `json:"user_id"`
	ChannelID   string `json:"channel_id"`
	Time        string `json:"time"`
	SubPlan     string `json:"sub_plan"`
	SubPlanName string `json:"sub_plan_name"`
	Months      int    `json:"months"`
	Context     string `json:"context"`
	SubMessage  struct {
		Message string         `json:"message"`
		Emotes  []*PUBSUBEmote `json:"emotes"`
	} `json:"sub_message"`
}

// NewPUBSUBSubscriptionMessage returns a new subscription message.
func NewPUBSUBSubscriptionMessage(message string) (*PUBSUBSubscriptionMessage, error) {
	var sub PUBSUBSubscriptionMessage

	// marshal message
	if err := json.Unmarshal([]byte(message), &sub); err != nil {
		return nil, err
	}

	// return sub
	return &sub, nil
}

// PUBSUBEmote is a twitch emote contained within a message.
type PUBSUBEmote struct {
	Start int `json:"start"`
	End   int `json:"end"`
	ID    int `json:"id"`
}

// PUBSUBBitsMessage is a message for a bits
// event on a pub sub message.
type PUBSUBBitsMessage struct {
	Data        *PUBSUBBitsMessageData `json:"data"`
	Version     string                 `json:"version"`
	MessageType string                 `json:"message_type"`
	MessageID   string                 `json:"message_id"`
}

// PUBSUBBitsMessageData is the data for a PUBSUBBitsMessage.
type PUBSUBBitsMessageData struct {
	UserName         string                             `json:"user_name"`
	ChannelName      string                             `json:"channel_name"`
	UserID           string                             `json:"user_id"`
	ChannelID        string                             `json:"channel_id"`
	Time             string                             `json:"time"`
	ChatMessage      string                             `json:"chat_message"`
	BitsUsed         int                                `json:"bits_used"`
	TotalBitsUsed    int                                `json:"total_bits_used"`
	Context          string                             `json:"context"`
	BadgeEntitlement *PUBSUBBitsMessageBadgeEntitlement `json:"badge_entitlement"`
}

// PUBSUBBitsMessageBadgeEntitlement contains badge entitlement
// meta data for PUBSUBBitsMessageData.
type PUBSUBBitsMessageBadgeEntitlement struct {
	NewVersion      int `json:"new_version"`
	PreviousVersion int `json:"previous_version"`
}

// NewPUBSUBBitsMessage returns a new bits message.
func NewPUBSUBBitsMessage(message string) (*PUBSUBBitsMessage, error) {
	var bits PUBSUBBitsMessage

	// marshal message
	if err := json.Unmarshal([]byte(message), &bits); err != nil {
		return nil, err
	}

	// return bits
	return &bits, nil
}

// PUBSUBCommerceMessage is a message for a commerce
// event on a pub sub message.
type PUBSUBCommerceMessage struct {
	UserName        string `json:"user_name"`
	DisplayName     string `json:"display_name"`
	ChannelName     string `json:"channel_name"`
	UserID          string `json:"user_id"`
	ChannelID       string `json:"channel_id"`
	Time            string `json:"time"`
	ItemImageURL    string `json:"item_image_url"`
	ItemDescription string `json:"item_description"`
	SupportsChannel bool   `json:"supports_channel"`
	PurchaseMessage struct {
		Message string         `json:"message"`
		Emotes  []*PUBSUBEmote `json:"emotes"`
	} `json:"purchase_message"`
}

// PUBSUBWhisperMessage is a message for a whisper
// event on a pub sub message.
type PUBSUBWhisperMessage struct{}
