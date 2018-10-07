package twitch

// PUBSUBMessageError is an error type on a pub sub message.
type PUBSUBMessageError string

const (
	errBadMessage PUBSUBMessageError = "ERR_BADMESSAGE"
	errBadAuth    PUBSUBMessageError = "ERR_BADAUTH"
	errBadTopic   PUBSUBMessageError = "ERR_BADTOPIC"
	errServer     PUBSUBMessageError = "ERR_SERVER"
	errServer2    PUBSUBMessageError = "Server Error"
)

func (e PUBSUBMessageError) String() string {
	return string(e)
}
