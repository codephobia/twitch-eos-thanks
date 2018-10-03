package main

import (
	"log"

	api "github.com/codephobia/twitch-eos-thanks/server/api"
	config "github.com/codephobia/twitch-eos-thanks/server/config"
	database "github.com/codephobia/twitch-eos-thanks/server/database"
	"github.com/codephobia/twitch-eos-thanks/server/twitch"
)

type Main struct {
	config   *config.Config
	database *database.Database
	api      *api.API
}

func main() {
	// make a new main
	err, _ := NewMain()

	if err != nil {
		log.Fatalf("[ERROR] main: %s", err)
	}
}

func NewMain() (*Main, error) {
	// load config
	c := config.NewConfig()
	if err := c.Load(); err != nil {
		return nil, err
	}

	// init database
	db := database.NewDatabase(c)
	if err := db.Init(); err != nil {
		return nil, err
	}

	// init twitch
	t := twitch.NewTwitch(c, db)
	if err := t.Init(); err != nil {
		return nil, err
	}

	// api
	api := api.NewAPI(c, db)
	if err := api.Init(); err != nil {
		return nil, err
	}

	// return main
	return &Main{
		config:   c,
		database: db,
		api:      api,
	}, nil
}
