package twitch

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
)

// hit codephobia api for follower list
func (t *Twitch) getApiResponse(url string) ([]byte, error) {
    // create http client
    client := &http.Client{}

    // create new request
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error generating request: %v", err)
    }
    
    // do get request
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error doing request: %v", err)
    }
    defer resp.Body.Close()

    // read body
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading body: %v", err)
    }

    return data, nil
}

// get a twitch response
func (t *Twitch) getTwitchResponse(version TwitchVersion, urlSuffix string) ([]byte, error) {
    // create http client
    client := &http.Client{}

    // build url with version prefix / suffix
    url := strings.Join([]string{version.Url(), urlSuffix}, "")
    
    // create new request
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error generating request: %v", err)
    }

    // add oauth token and client id to headers
    req.Header.Add("Authorization", t.config.TwitchOAuthToken)
    req.Header.Add("Client-ID", t.config.TwitchClientID)

    // add accept header for V5 api calls
    if version == TwitchV5 {
        req.Header.Add("Accept", "application/vnd.twitchtv.v5+json")
    }
    
    // do get request
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error doing request: %v", err)
    }
    defer resp.Body.Close()

    // read body
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading body: %v", err)
    }

    return data, nil
}