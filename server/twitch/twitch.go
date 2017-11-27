package twitch

import(
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"
    
    "encoding/base64"
    
    config   "github.com/codephobia/twitch-eos-thanks/server/config"
    database "github.com/codephobia/twitch-eos-thanks/server/database"
    util     "github.com/codephobia/twitch-eos-thanks/server/util"
)

var (
    TWITCH_API_URL       string = "https://api.twitch.tv/helix"
    TWITCH_USERS_URL     string = "/users?"
    TWITCH_FOLLOWERS_URL string = "/users/follows?to_id="
    
    TWITCH_DB_BUCKET          []string = []string{"twitch"}
    TWITCH_FOLLOWER_DB_BUCKET []string = append(TWITCH_DB_BUCKET, "followers")
)

type Twitch struct {
    config   *config.Config
    database *database.Database
    timer    *util.Timer
        
    cursor    string
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
func NewTwitch(c *config.Config, db *database.Database) (*Twitch, error) {
    var cursor database.FollowerCursor

    // init the bucket
    if err := db.InitBucket(TWITCH_FOLLOWER_DB_BUCKET); err != nil {
        return nil, fmt.Errorf("init twitch bucket: ", err)
    }
    
    // load the cursor from the database if it exists
    err, curBytes := db.Get(TWITCH_DB_BUCKET, "cursor")
    if err != nil {
        return nil, fmt.Errorf("unable to load cursor from db: %s", err)
    }
    
    // unmarshal cursor
    if len(curBytes) > 0 {
        if err := json.Unmarshal(curBytes, &cursor); err != nil {
            return nil, fmt.Errorf("unable to unmarshal cursor from db: %s", err)
        }
    }
    
    // return new twitch struct
    return &Twitch{
        config:   c,
        database: db,
        
        cursor:   cursor.Cursor,
    }, nil
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
    
    // run timer to poll twitch api
    t.startTimer()
    
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
    log.Printf("[INFO] getFollowers: checking twitch for followers")
    
    limit := 30
    loop := true
    
    for loop {
        // build out url
        u := []string{TWITCH_API_URL, TWITCH_FOLLOWERS_URL, t.config.TwitchChannelID, "&first=", strconv.Itoa(limit)}
        
        // check for cursor
        if len(t.cursor) > 0 {
            u = append(u, strings.Join([]string{"&after=", t.cursor}, ""))
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
        
        // save cursor for next loop, if we got results
        if len(followerResp.Data) > 0 {
            t.cursor = followerResp.Pagination.Cursor
            log.Printf("[INFO] getFollowers: updating cursor: %s", t.cursor)
            
            base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(t.cursor)))
            l, _ := base64.StdEncoding.Decode(base64Text, []byte(t.cursor))
            log.Printf("base64: %s\n", base64Text[:l])
        }
        
        // check if we need to keep looping
        cnt := len(followerResp.Data)
        if (cnt < limit) {
            // stop loop
            loop = false
        }

        // sleep so we don't hammer twitch api
        time.Sleep(2 * time.Second)
    }

    log.Printf("[INFO] getFollowers: found [%d] new followers", len(t.Followers))
    
    // save the cursor to the database for next launch
    t.database.Put(TWITCH_DB_BUCKET, "cursor", &database.FollowerCursor{
        Cursor: t.cursor,
    })
    
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
    // check if we found followers
    if len(t.Followers) == 0 {
        return nil
    }
    
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
    // check if we found followers
    if len(t.Followers) == 0 {
        return nil
    }
    
    for _, follower := range t.Followers {
        // convert follower for db
        dbFollower := &database.Follower{
            ID:              follower.FollowerID,
            FollowedAt:      follower.FollowedAt,
            DisplayName:     follower.UserData.DisplayName,
            ProfileImageUrl: follower.UserData.ProfileImageUrl,
        }
        
        // put the follower data
        if err := t.database.Put(TWITCH_FOLLOWER_DB_BUCKET, dbFollower.ID, dbFollower); err != nil {
            return fmt.Errorf("saving follower [%s]: %+v", err)
        }
    }
    
    // reset the followers
    t.Followers = make([]*Follower, 0)
    
    return nil
}

// start a timer for twitch api polling
func (t *Twitch) startTimer() {
    d := 5 * time.Second
    t.timer = util.NewTimer(d, false, t.cron)
}

// cron function run by timer
func (t *Twitch) cron() {
    // get twitch followers
    err := t.Get()
    if err != nil {
        log.Printf("twitch cron: %s", err)
    }
}