package twitch

import "encoding/json"

// PUBSUBRequest is a twitch pub sub request.
type PUBSUBRequest struct {
	Type string             `json:"type"`
	Data *PUBSUBRequestData `json:"data"`
}

// PUBSUBRequestData is the payload on a twitch pub sub request.
type PUBSUBRequestData struct {
	Topics    []string `json:"topics"`
	AuthToken string   `json:"auth_token"`
}

// NewPUBSUBRequest returns a new pub sub request.
func NewPUBSUBRequest(requestType string, topics []string, authToken string) *PUBSUBRequest {
	return &PUBSUBRequest{
		Type: requestType,
		Data: &PUBSUBRequestData{
			Topics:    topics,
			AuthToken: authToken,
		},
	}
}

// ToBytes returns a pub sub request as an array of bytes.
func (r *PUBSUBRequest) ToBytes() ([]byte, error) {
	// marshal request in to bytes
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	// return request bytes
	return data, nil
}
