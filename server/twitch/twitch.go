package twitch

import(
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "time"
    
    config   "github.com/codephobia/twitch-eos-thanks/server/config"
    database "github.com/codephobia/twitch-eos-thanks/server/database"
)

var (
    TWITCH_API_URL       string = "https://api.twitch.tv/helix"
    TWITCH_USERS_URL     string = "/users?id="
    TWITCH_FOLLOWERS_URL string = "/users/follows?to_id="
)

type Twitch struct {
    config   *config.Config
    database *database.Database
    
    Followers []Follower
}

type FollowersResp struct {
    Data       []Follower `json:"data"`
    Pagination struct {
        Cursor string
    }
}

type Follower struct {
    FollowerID string `json:"from_id"`
    FollowedAt string `json:"followed_at"`
    UserData   TwitchUser
}

type UserResp struct {
    Data []TwitchUser `json:"data"`
}

type TwitchUser struct {
    User struct {
        DisplayName     string `json:"display_name"`
        ProfileImageUrl string `json:"profile_image_url"`        
    } `json:"user"`
}

// create twitch
func NewTwitch(config *config.Config, database *database.Database) *Twitch {
    return &Twitch{
        config:   config,
        database: database,
    }
}

// get data from helix
func (t *Twitch) Get() error {
    // get follower ids
    if err := t.getFollowers(); err != nil {
        return err
    }
    
    // get user data from follower ids
    if err := t.getFollowerUserData(); err != nil {
        return err
    }
    
    // save the followers to the database
    
    log.Printf("followers: %+v", t.Followers)
    
    return nil
}

func (t *Twitch) getTwitchResponse(url string) ([]byte, error) {
    // create http client
    client := &http.Client{}

    // create new request
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error generating request: %v", err)
    }

    // add client id to headers
    req.Header.Add("Client-ID", t.config.TwitchClientID)

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

func (t *Twitch) getFollowers() error {
    cursor := ""
    limit := 100
    loop := true
    
    for loop {
        // build out url
        u := []string{TWITCH_API_URL, TWITCH_FOLLOWERS_URL, t.config.TwitchChannelID, "&first=", "100"}
        
        // check for cursor
        if len(cursor) > 0 {
            u = append(u, strings.Join([]string{"&after=", cursor}, ""))
        }
        url := strings.Join(u, "")

        // get followers from twitch
        body, err := t.getTwitchResponse(url)
        if err != nil {
            return err
        }
        
        // decode body
        followerResp := &FollowersResp{}
        if err := json.NewDecoder(bytes.NewReader(body)).Decode(followerResp); err != nil {
            return fmt.Errorf("body decode: ", err)
        }    

        // save followers to twitch struct
        t.Followers = append(t.Followers, followerResp.Data...)
        
        // save cursor for next loop
        cursor = followerResp.Pagination.Cursor

        // check if we need to keep looping
        cnt := len(followerResp.Data)
        if (cnt < limit) {
            // stop loop
            loop = false
        }

        // sleep so we don't hammer twitch api
        time.Sleep(2 * time.Second)
    }
    
    return nil
}

func (t *Twitch) getFollowerUserData() error {
    // loop through followers
    for _, follower := range t.Followers {
        // build out url
        u := []string{TWITCH_API_URL, TWITCH_USERS_URL, follower.FollowerID}
        url := strings.Join(u, "")
        
        // get user data from twitch
        body, err := t.getTwitchResponse(url)
        if err != nil {
            return err
        }
        
        log.Printf("body: %+v", string(body))
        
        // decode body
        userResp := &UserResp{}
        if err := json.NewDecoder(bytes.NewReader(body)).Decode(userResp); err != nil {
            return fmt.Errorf("body decode: ", err)
        }
        
        // make sure we got data
        if len(userResp.Data) == 0 {
            return fmt.Errorf("no data: ", string(body))
        }
        
        // update follower user data
        follower.UserData = userResp.Data[0]
        
        log.Printf("user: %+v", userResp.Data)
        
        // sleep so we don't hammer twitch api
        time.Sleep(2 * time.Second)
    }
    
    return nil
}