package twitch

import (
    "bytes"
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

var (
    TWITCH_V5_STREAMS_URL string = "/streams"
)

// get the time of the current stream start
func (t *Twitch) getStreamStart() (time.Time, error) {
    // make default time
    tm := time.Unix(0, 0)
    
    // build out url
    u := []string{TWITCH_V5_STREAMS_URL, "/", t.config.TwitchChannelID, "?stream_type=live"}
    url := strings.Join(u, "")

    // get current stream
    body, err := t.getTwitchResponse(TwitchV5, url)
    if err != nil {
        return tm, err
    }
    
    // decode body
    streamsResp := &StreamsResp{}
    if err := json.NewDecoder(bytes.NewReader(body)).Decode(streamsResp); err != nil {
        return tm, fmt.Errorf("body decode: ", err)
    }
    
    // if we have a stream time
    if len(streamsResp.Stream.CreatedAt) > 0 {
        // convert string to time
        streamTime, err := time.Parse(time.RFC3339, streamsResp.Stream.CreatedAt)
        if err != nil {
            return tm, err
        }
        
        // update time
        tm = streamTime
    }
    
    // return time
    return tm, nil
}