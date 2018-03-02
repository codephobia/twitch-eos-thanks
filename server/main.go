package main

import (
	"log"

	api "github.com/codephobia/twitch-eos-thanks/server/api"
	config "github.com/codephobia/twitch-eos-thanks/server/config"
	database "github.com/codephobia/twitch-eos-thanks/server/database"
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

	// api
	api := api.NewAPI(c, db)
	if err := api.Init(); err != nil {
		return err, nil
	}

	// return main
	return nil, &Main{
		config:   c,
		database: db,
		api:      api,
	}
}
