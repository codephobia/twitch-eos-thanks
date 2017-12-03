package database

import (
    "encoding/json"
    "fmt"
    "path"
    "time"
    
    bolt "github.com/boltdb/bolt"
    
    config "github.com/codephobia/twitch-eos-thanks/app/config"
)

var (
    DB_PATH string = "."
)

type Database struct {
    config *config.Config
    
    boltDB *bolt.DB
}

// create new database
func NewDatabase(config *config.Config) *Database {
    return &Database{
        config: config,
    }
}

// init database
func (db *Database) Init() error {
    // create the bolt database
    boltDB, err := bolt.Open(path.Join(DB_PATH, db.config.DBFileName), 0600, nil)
    if err != nil {
        return fmt.Errorf("database init: ", err)
    }
    
    // set db
    db.boltDB = boltDB
    
    return nil
}

// init a bucket
func (db *Database) InitBucket(buckets []string) error {
    // make sure we have buckets
    if len(buckets) == 0 {
        return fmt.Errorf("init bucket: bucket required")
    }
    
    // tx
    return db.boltDB.Update(func(tx *bolt.Tx) error {
        var bkt *bolt.Bucket
        
        for i, bucket := range buckets {
            // if first bucket, create it on tx
            if i == 0 {
                newBkt, err := tx.CreateBucketIfNotExists([]byte(bucket))
                bkt = newBkt
                
                if (err != nil) {
                    return fmt.Errorf("error creating bucket [%s]: %v", bucket, err)
                }
            } else {
                newBkt, err := bkt.CreateBucketIfNotExists([]byte(bucket))
                bkt = newBkt
                
                if (err != nil) {
                    return fmt.Errorf("error creating bucket [%s]: %v", bucket, err)
                }
            }
        }
        
        return nil
    })
}

// put entry
func (db *Database) Put(buckets []string, key string, value interface{}) error {
    // make sure we have buckets
    if len(buckets) == 0 {
        return fmt.Errorf("put: bucket required")
    }
    
    // tx
    return db.boltDB.Update(func(tx *bolt.Tx) error {
        var bkt *bolt.Bucket
        
        k := []byte(key)
        v, _ := json.Marshal(value)
        
        // iterate through the buckets
        for i, bucket := range buckets {
            // if first bucket, load from tx
            if i == 0 {
                bkt = tx.Bucket([]byte(bucket))
            } else {
                bkt = bkt.Bucket([]byte(bucket))
            }
            
            // check for errors
            if bkt == nil {
                return fmt.Errorf("put: error selecting bucket: %s", bucket)
            }
        }
        
        // put into bucket
        return bkt.Put(k, v)
    })
}

// deep get entry
func (db *Database) Get(buckets []string, key string) (error, []byte) {
    var data []byte
    
    // make sure we have buckets
    if len(buckets) == 0 {
        return fmt.Errorf("get: bucket required"), data
    }
    
    err := db.boltDB.View(func(tx *bolt.Tx) error {
        var bkt *bolt.Bucket
        
        k := []byte(key)
        
        // iterate through the buckets
        for i, bucket := range buckets {
            // if first bucket, load from tx
            if i == 0 {
                bkt = tx.Bucket([]byte(bucket))
            } else {
                bkt = bkt.Bucket([]byte(bucket))
            }
            
            // check for errors
            if bkt == nil {
                return fmt.Errorf("get: error selecting bucket: %s", bucket)
            }
        }
        
        // get data
        data = bkt.Get(k)
        
        return nil
    })
    
    return err, data
}

// deep get array
func (db *Database) GetAll(buckets []string) (error, [][]byte) {
    data := make([][]byte, 0)
    
    // make sure we have buckets
    if len(buckets) == 0 {
        return fmt.Errorf("get all: bucket required"), data
    }
    
    // tx
    err := db.boltDB.View(func(tx *bolt.Tx) error {
        var bkt *bolt.Bucket
        
        // iterate through the buckets
        for i, bucket := range buckets {
            // if first bucket, load from tx
            if i == 0 {
                bkt = tx.Bucket([]byte(bucket))
            } else {
                bkt = bkt.Bucket([]byte(bucket))
            }
            
            // check for errors
            if bkt == nil {
                return fmt.Errorf("get all: error selecting bucket: %s", bucket)
            }
        }
        
        c := bkt.Cursor()
        
        // loop through items in db
        for k, v := c.Last(); k != nil; k, v = c.Prev() {
            // append to data
            data = append(data, v)
        }
        
        return nil
    })
    
    return err, data
}

// deep get array since time
func (db *Database) GetAllSince(buckets []string, since time.Time, timeKey string) (error, [][]byte) {
    data := make([][]byte, 0)
    
    // make sure we have buckets
    if len(buckets) == 0 {
        return fmt.Errorf("get all since: bucket required"), data
    }
    
    // tx
    err := db.boltDB.View(func(tx *bolt.Tx) error {
        var bkt *bolt.Bucket
        
        // iterate through the buckets
        for i, bucket := range buckets {
            // if first bucket, load from tx
            if i == 0 {
                bkt = tx.Bucket([]byte(bucket))
            } else {
                bkt = bkt.Bucket([]byte(bucket))
            }
            
            // check for errors
            if bkt == nil {
                return fmt.Errorf("get all since: error selecting bucket: %s", bucket)
            }
        }
        
        c := bkt.Cursor()
        
        // loop through items in db
        for k, v := c.Last(); k != nil; k, v = c.Prev() {
            var entry map[string]interface{}
            
            // unmarshal the entry
            if err := json.Unmarshal(v, &entry); err != nil {
                return fmt.Errorf("get all since: error unmarshalling entry: %s", err)
            }
            
            // convert entry string to time
            entryTime, err := time.Parse(time.RFC3339, entry[timeKey].(string))
            if err != nil {
                return fmt.Errorf("get all since: error converting time: %s", err)
            }
            
            // if time is before since
            if entryTime.After(since) {
                // append to data
                data = append(data, v)
            }
            
        }
        
        return nil
    })
    
    return err, data
}

// return count of keys in bucket
func (db *Database) Count(buckets []string) (int, error) {
    // default count to zero
    count := 0
    
    // make sure we have buckets
    if len(buckets) == 0 {
        return count, fmt.Errorf("count: bucket required")
    }
    
    // tx
    err := db.boltDB.View(func(tx *bolt.Tx) error {
        var bkt *bolt.Bucket
        
        // iterate through the buckets
        for i, bucket := range buckets {
            // if first bucket, load from tx
            if i == 0 {
                bkt = tx.Bucket([]byte(bucket))
            } else {
                bkt = bkt.Bucket([]byte(bucket))
            }
            
            // check for errors
            if bkt == nil {
                return fmt.Errorf("count: error selecting bucket: %s", bucket)
            }
        }
        
        // set count
        count = bkt.Stats().KeyN
        
        return nil
    })
    
    return count, err
}