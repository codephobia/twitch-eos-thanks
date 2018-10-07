package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	tokenURL = "https://id.twitch.tv/oauth2/token"
)

// RefreshTokenResp is a successful response from a refresh request.
type RefreshTokenResp struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
}

// RefreshTokenErrorResp is an invalid response from a refresh request.
type RefreshTokenErrorResp struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// refresh the access token from twitch
func (p *PUBSUB) refreshToken() error {
	// send request to twitch for refresh token update
	newAccessToken, err := p.requestRefreshToken()
	if err != nil {
		return fmt.Errorf("refresh token: %s", err)
	}

	// update config
	p.config.TwitchChannelOAuthToken = newAccessToken

	// save updated config to file
	return p.config.Save()
}

// request new access token using refresh token
func (p *PUBSUB) requestRefreshToken() (string, error) {
	// build url with version prefix / suffix
	url := strings.Join([]string{
		tokenURL,
		"?grant_type=refresh_token",
		"&refresh_token=" + p.config.TwitchChannelRefreshToken,
		"&client_id=" + p.config.TwitchClientID,
		"&client_secret=" + p.config.TwitchClientSecret,
	}, "")

	// create new request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("error generating request: %v", err)
	}

	// do post request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing request: %v", err)
	}
	defer resp.Body.Close()

	// check for expected status codes
	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusBadRequest) {
		return "", fmt.Errorf("invalid response code: %d", resp.StatusCode)
	}

	// read body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %v", err)
	}

	// error status code
	if resp.StatusCode == http.StatusBadRequest {
		var errorResp RefreshTokenErrorResp

		// unmarshal data
		if err := json.Unmarshal(data, &errorResp); err != nil {
			return "", fmt.Errorf("error unmarshalling bad request: %s", err)
		}

		// return response error
		return "", fmt.Errorf("error reading body: %s: %s", errorResp.Error, errorResp.Message)
	}

	// success status code
	var successResp RefreshTokenResp

	// unmarshal data
	if err := json.Unmarshal(data, &successResp); err != nil {
		return "", fmt.Errorf("error unmarshalling successful request: %s", err)
	}

	return successResp.AccessToken, nil
}
