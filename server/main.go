package main

import (
    "log"
    
    config   "github.com/codephobia/twitch-eos-thanks/server/config"
    database "github.com/codephobia/twitch-eos-thanks/server/database"
    twitch   "github.com/codephobia/twitch-eos-thanks/server/twitch"
    api      "github.com/codephobia/twitch-eos-thanks/server/api"
)

type Main struct {
    config   *config.Config
    database *database.Database
    twitch   *twitch.Twitch
    api      *api.Api
}

func main() {
    // make a new main
    err, _ := NewMain()
    
    if err != nil {
        log.Fatal("[ERROR] main: ", err)
    }
}

func NewMain() (error, *Main) {
    // load config
    c := config.NewConfig()
    if err := c.Load(); err != nil {
        return err, nil
    }
    
    // init database
    db := database.NewDatabase(c)
    if err := db.Init(); err != nil {
        return err, nil
    }
    
    // twitch
    twitch := twitch.NewTwitch(c, db)
    if err := twitch.Get(); err != nil {
        return err, nil
    }
    
    // return main
    return nil, &Main{
        config:   c,
        database: db,
        twitch:   twitch,
    }
}