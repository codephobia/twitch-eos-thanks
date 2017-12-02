package twitch

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"
    
    database "github.com/codephobia/twitch-eos-thanks/app/database"
)

type Follower struct {
    FollowerID string `json:"from_id"`
    FollowedAt string `json:"followed_at"`
    UserData   *TwitchUser
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

func (t *Twitch) getFollowers() error {
    log.Printf("[INFO] getFollowers: checking api for followers")

    // get latest cached follow time
    latestFollowTime, err := t.getLatestFollowerTime()
    if err != nil {
        return fmt.Errorf("latest follower time: ", err)
    }
    
    i := 0
    loop := true

    for loop {
        offset := i * TWITCH_API_FOLLOWER_LIMIT
        
        // build out url
        u := []string{"http://", t.config.CodephobiaApiHost, ":", t.config.CodephobiaApiPort, "/followers?channelID=", t.config.TwitchChannelID, "&latest=", strconv.FormatInt(latestFollowTime.UnixNano(), 10), "&limit=", strconv.Itoa(TWITCH_API_FOLLOWER_LIMIT)}
        
        // set offset if not on first page
        if i > 0 {
            u = append(u, strings.Join([]string{"&offset=", strconv.Itoa(offset)}, ""))
        }
        url := strings.Join(u, "")

        // get followers from twitch
        body, err := t.getApiResponse(url)
        if err != nil {
            return err
        }
        
        // decode body
        followerResp := &FollowersResp{}
        if err := json.NewDecoder(bytes.NewReader(body)).Decode(followerResp); err != nil {
            return fmt.Errorf("body decode: ", err)
        }

        // loop through response followers
        for _, follower := range followerResp.Data {
            // save followers to twitch struct
            newFollower := &Follower{
                FollowerID: follower.FollowerID,
                FollowedAt: follower.Timestamp,
            }
            t.Followers = append(t.Followers, newFollower)            
        }
        
        // check if we need to keep looping
        cnt := len(followerResp.Data)
        if (cnt < TWITCH_API_FOLLOWER_LIMIT) {
            // stop loop
            loop = false
        }

        // sleep so we don't hammer api
        time.Sleep(TWITCH_API_DELAY)
    }

    log.Printf("[INFO] getFollowers: found [%d] new followers", len(t.Followers))
    
    return nil
}

func (t *Twitch) getLatestFollowerTime() (time.Time, error) {
    lt := time.Unix(0, 0)
    
    // get follower count
    count, err := t.database.Count(TWITCH_FOLLOWER_DB_BUCKET)
    if err != nil {
        return lt, fmt.Errorf("count: ", err)
    }
    
    // if we have cached followers, get latest follow time
    if count > 0 {
        // get all followers from database
        err, dbFollowers := t.database.GetAll(TWITCH_FOLLOWER_DB_BUCKET)
        if err != nil {
            return lt, fmt.Errorf("get followers: ", err)
        }
        
        // loop through followers
        for _, dbFollower := range dbFollowers {
            // unmarshal follower
            var follower database.Follower
            if err := json.Unmarshal(dbFollower, &follower); err != nil {
                return lt, fmt.Errorf("unmarshal follower: ", err)
            }

            // parse date
            ft, _ := time.Parse(time.RFC3339, follower.FollowedAt)
            
            // check if follower date is more recent
            if ft.After(lt) {
                lt = ft
            }
        }
    }
    
    return lt, nil
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
    
    for loop {
        // url ids
        var urlIds []string

        // create slice of ids for loop
        start := i * TWITCH_API_USER_LIMIT
        end := start + TWITCH_API_USER_LIMIT
        
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
        u := []string{TWITCH_HELIX_USERS_URL, urlIdsString}
        url := strings.Join(u, "")
        
        // get user data from twitch
        body, err := t.getTwitchResponse(TwitchHelix, url)
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
        time.Sleep(TWITCH_API_DELAY)
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