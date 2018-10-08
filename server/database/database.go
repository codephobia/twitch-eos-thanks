package database

import (
	"fmt"
	"strings"

	mgo "gopkg.in/mgo.v2"

	config "github.com/codephobia/twitch-eos-thanks/server/config"
)

const (
	collectionFollowers   = "followers"
	collectionSubscribers = "subscribers"
	collectionBits        = "bits"
)

// Database handles the MongoDB connection.
type Database struct {
	config *config.Config

	session     *mgo.Session
	database    *mgo.Database
	followers   *mgo.Collection
	subscribers *mgo.Collection
	bits        *mgo.Collection
}

// NewDatabase returns a new database.
func NewDatabase(c *config.Config) *Database {
	return &Database{
		config: c,
	}
}

// Init initializes a new database.
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

	// followers
	db.initFollowers()

	// subscribers
	db.initSubscribers()

	// bits
	db.initBits()
}

// init followers collection
func (db *Database) initFollowers() {
	db.followers = db.database.C(collectionFollowers)
}

// init subscribers collection
func (db *Database) initSubscribers() {
	db.subscribers = db.database.C(collectionSubscribers)
}

// init bits collection
func (db *Database) initBits() {
	db.bits = db.database.C(collectionBits)
}
