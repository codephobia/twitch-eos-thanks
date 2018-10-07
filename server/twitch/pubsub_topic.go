package twitch

// PUBSUBTopic is a twitch pub sub topic.
type PUBSUBTopic string

const (
	PUBSUBTopicSubscription PUBSUBTopic = "channel-subscribe-events-v1"
	PUBSUBTopicBits         PUBSUBTopic = "channel-bits-events-v1"
	PUBSUBTopicCommerce     PUBSUBTopic = "channel-commerce-events-v1"
	PUBSUBTopicWhispers     PUBSUBTopic = "whispers"
)

func (t PUBSUBTopic) String() string {
	return string(t)
}
