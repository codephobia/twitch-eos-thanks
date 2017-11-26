package twitch

import(
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "time"
    
    config   "github.com/codephobia/twitch-eos-thanks/server/config"
    database "github.com/codephobia/twitch-eos-thanks/server/database"
)

var (
    TWITCH_API_URL       string = "https://api.twitch.tv/helix"
    TWITCH_USERS_URL     string = "/users?"
    TWITCH_FOLLOWERS_URL string = "/users/follows?to_id="
    
    TWITCH_FOLLOWER_DB_BUCKET []string = []string{"twitch", "followers"}
)

type Twitch struct {
    config   *config.Config
    database *database.Database
    
    Followers []*Follower
}

type FollowersResp struct {
    Data       []*Follower `json:"data"`
    Pagination struct {
        Cursor string
    }
}

type Follower struct {
    FollowerID string `json:"from_id"`
    FollowedAt string `json:"followed_at"`
    UserData   *TwitchUser
}

type UserResp struct {
    Data []*TwitchUser `json:"data"`
}

type TwitchUser struct {
    ID              string `json:"id"`
    DisplayName     string `json:"display_name"`
    ProfileImageUrl string `json:"profile_image_url"`        
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
    if err := t.saveFollowers(); err != nil {
        return err
    }
    
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
        u := []string{TWITCH_API_URL, TWITCH_FOLLOWERS_URL, t.config.TwitchChannelID, "&first=", strconv.Itoa(limit)}
        
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

// find a twitch follower by id
func (t *Twitch) findFollowerById(id string) (*Follower, error) {
    for _, v := range t.Followers {
        if v.FollowerID == id {
            return v, nil
        }
    }
    
    return nil, fmt.Errorf("unable to find follower: %s", id)
}

func (t *Twitch) getFollowerUserData() error {
    // loop through followers to get ids
    var ids []string
    for _, follower := range t.Followers {
        ids = append(ids, follower.FollowerID)
    }
    
    // loop through ids to get user data
    loop := true
    i := 0
    limit := 100
    
    for loop {
        // url ids
        var urlIds []string

        // create slice of ids for loop
        start := i * limit
        end := start + limit
        
        // check end to make sure it isn't larger than ids array
        if end > len(ids) {
            end = len(ids)
        }
        
        // current loop ids
        curIds := ids[start:end]
        
        // loop through ids to build out url strings
        for _, id := range curIds {
            urlIds = append(urlIds, strings.Join([]string{"id=", id}, ""))
        }
        
        // combine url ids
        urlIdsString := strings.Join(urlIds, "&")
        
        // build out url
        u := []string{TWITCH_API_URL, TWITCH_USERS_URL, urlIdsString}
        url := strings.Join(u, "")
        
        // get user data from twitch
        body, err := t.getTwitchResponse(url)
        if err != nil {
            return err
        }
        
        // decode body
        userResp := &UserResp{}
        if err := json.NewDecoder(bytes.NewReader(body)).Decode(userResp); err != nil {
            return fmt.Errorf("body decode: ", err)
        }
        
        // make sure we got data
        if len(userResp.Data) == 0 {
            return fmt.Errorf("no data: ", string(body))
        }
        
        // loop through response users and assign to followers
        for _, userData := range userResp.Data {
            follower, err := t.findFollowerById(userData.ID)
            if err != nil {
                return err
            }
            
            follower.UserData = userData
        }
        
        // check if we need to stop looping
        if end >= len(ids) {
            loop = false
        }
        
        // increase i
        i++

        // sleep so we don't hammer twitch api
        time.Sleep(2 * time.Second)
    }
    
    return nil
}

// the the followers to the database
func (t *Twitch) saveFollowers() error {
    // init the bucket
    if err := t.database.InitBucket(TWITCH_FOLLOWER_DB_BUCKET); err != nil {
        return fmt.Errorf("saving followers: ", err)
    }
    
    for _, follower := range t.Followers {
        // put the follower data
        if err := t.database.Put(TWITCH_FOLLOWER_DB_BUCKET, follower.FollowerID, follower); err != nil {
            return fmt.Errorf("saving follower [%s]: %+v", err)
        }
    }
    
    return nil
}