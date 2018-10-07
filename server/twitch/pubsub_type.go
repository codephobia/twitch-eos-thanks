package twitch

// PUBSUBType is a type of pub sub.
type PUBSUBType string

const (
	PUBSUBTypeResponse  PUBSUBType = "RESPONSE"
	PUBSUBTypePing      PUBSUBType = "PING"
	PUBSUBTypePong      PUBSUBType = "PONG"
	PUBSUBTypeReconnect PUBSUBType = "RECONNECT"
	PUBSUBTypeListen    PUBSUBType = "LISTEN"
	PUBSUBTypeUnListen  PUBSUBType = "UNLISTEN"
	PUBSUBTypeMessage   PUBSUBType = "MESSAGE"
)

func (t PUBSUBType) String() string {
	return string(t)
}
