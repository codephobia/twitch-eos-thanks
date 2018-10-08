package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	ps "github.com/mitchellh/go-ps"

	api "github.com/codephobia/twitch-eos-thanks/app/api"
	config "github.com/codephobia/twitch-eos-thanks/app/config"
	database "github.com/codephobia/twitch-eos-thanks/app/database"
	twitch "github.com/codephobia/twitch-eos-thanks/app/twitch"
)

type Main struct {
	config   *config.Config
	database *database.Database
	twitch   *twitch.Twitch
	api      *api.Api
}

func main() {
	// make sure app isn't already running
	if err := checkRunning(); err != nil {
		log.Fatal(err)
	}

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

	// twitch
	twitch, err := twitch.NewTwitch(c, db)
	if err != nil {
		return err, nil
	}
	if err := twitch.Get(); err != nil {
		return err, nil
	}

	// api
	api := api.NewApi(c, db, twitch)
	if err := api.Init(); err != nil {
		return err, nil
	}

	// return main
	return nil, &Main{
		config:   c,
		database: db,
		twitch:   twitch,
		api:      api,
	}
}

// checks if the app is already running
func checkRunning() error {
	// get filename of app
	_, fileName := filepath.Split(os.Args[0])

	// get process list
	processes, err := ps.Processes()
	if err != nil {
		return err
	}

	// check if more than one of app running
	cnt := 0
	for _, p := range processes {
		if p.Executable() == fileName {
			cnt++
		}
	}

	// if app running more than once, error
	if cnt > 1 {
		return fmt.Errorf("App is already running")
	}

	// return no error
	return nil
}
