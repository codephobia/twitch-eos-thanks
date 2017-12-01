package database

import (
    "fmt"
    
    "gopkg.in/mgo.v2/bson"
)

type Follower struct {
    ID         bson.ObjectId `bson:"_id,omitempty" json:"ID,omitempty"`
    ChannelID  string        `bson:"channel_id,omitempty" json:"channelID,omitempty"`
    FollowerID string        `bson:"follower_id,omitempty" json:"followerID,omitempty"`
    Timestamp  string        `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
}

// add follower
func (db *Database) AddFollower(f *Follower) error {
    // check if follower already exists
    followed, err := db.hasFollower(f.ChannelID, f.FollowerID)
    if err != nil {
        return err
    }
    
    // skip adding follower to database if they are already following
    if followed {
        return fmt.Errorf("found duplicate follower [%s] for channel [%s]", f.FollowerID, f.ChannelID)
    }
    
    // insert new follower
    return db.followers.Insert(f)
}

// check for follower
func (db *Database) hasFollower(channelID string, followerID string) (bool, error) {
    // build query
    query := db.followers.Find(bson.M{
        "channel_id": channelID,
        "follower_id": followerID,
    })
    
    // get count
    count, err := query.Count()
    if err != nil {
        return false, err
    }
    
    // return that we found a follower
    if count > 0 {
        return true, nil
    }
    
    // return that we didn't find a follower
    return false, nil
}

// remove follower
func (db *Database) RemoveFollower(f *Follower) error {
    // remove the follower from the database
    return db.followers.Remove(bson.M{
        "channel_id": f.ChannelID,
        "follower_id": f.FollowerID,
    })
}

// get followers
func (db *Database) GetFollowers(channelID string, limit int, offset int) ([]*Follower, error) {
    followers := make([]*Follower, 0)
    
    // build query
    query := db.followers.Find(bson.M{
        "channel_id": channelID,
    })
    
    // add filters
    query.Limit(limit).Skip(offset).Sort("-timestamp").Select(bson.M{
        "_id": 0,
        "follower_id": 1,
        "timestamp":   1,
    })
    
    // get followers
    err := query.All(&followers)
    if err != nil {
        return followers, fmt.Errorf("unable to get followers: %s", err)
    }
    
    return followers, nil
}