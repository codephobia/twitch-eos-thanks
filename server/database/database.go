package database

import (
    "fmt"
    "strings"
    
    mgo "gopkg.in/mgo.v2"
    
    config "github.com/codephobia/twitch-eos-thanks/server/config"
)

var (
    COLLECTION_FOLLOWERS string = "followers"
)

type Database struct {
    config *config.Config
    
    session   *mgo.Session
    database  *mgo.Database
    followers *mgo.Collection
}

func NewDatabase(c *config.Config) *Database {
    return &Database{
        config: c,
    }
}

func (db *Database) Init() error {
    // create mongo session
    mongoDBUrl := strings.Join([]string{db.config.MongoDBHost, db.config.MongoDBPort}, ":")
    session, err := mgo.Dial(mongoDBUrl)
    if err != nil {
        return fmt.Errorf("unable to dial server [%s]: %s", mongoDBUrl, err)
    }
    
    // store session
    db.session = session
    db.session.SetMode(mgo.Monotonic, true)
    
    // init followers
    db.initDatabase()
    
    return nil
}

// init database
func (db *Database) initDatabase() {
    db.database = db.session.DB(db.config.MongoDBDatabase)
    
    db.initFollowers()
}

// init followers collection
func (db *Database) initFollowers() {
    db.followers = db.database.C(COLLECTION_FOLLOWERS)
}