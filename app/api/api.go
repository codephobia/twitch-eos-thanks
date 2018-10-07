package api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	config "github.com/codephobia/twitch-eos-thanks/app/config"
	database "github.com/codephobia/twitch-eos-thanks/app/database"
	twitch "github.com/codephobia/twitch-eos-thanks/app/twitch"
)

type Api struct {
	config   *config.Config
	database *database.Database
	twitch   *twitch.Twitch

	server *http.Server
}

func NewApi(config *config.Config, database *database.Database, twitch *twitch.Twitch) *Api {
	return &Api{
		config:   config,
		database: database,
		twitch:   twitch,
	}
}

func (api *Api) Init() error {
	// create the server
	api.server = &http.Server{
		Handler:      handlers.CompressHandler(handlers.CORS()(api.Handler())),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// create a listener
	hostUrl := strings.Join([]string{api.config.ApiHost, ":", api.config.ApiPort}, "")
	listener, err := net.Listen("tcp", hostUrl)
	if err != nil {
		return fmt.Errorf("error starting api server: %s", err)
	}

	// run server
	log.Printf("API Server running: %s", listener.Addr().String())
	api.server.Serve(listener)

	return nil
}

func (api *Api) Handler() http.Handler {
	// create router
	r := mux.NewRouter()

	// check
	r.Handle("/check", api.handleCheck())

	// settings
	r.Handle("/settings", api.handleSettings())

	// followers
	r.Handle("/followers", api.handleFollowers())

	// followers
	r.Handle("/subscribers", api.handleSubscribers())

	// shutdown
	r.Handle("/shutdown", api.handleShutdown())

	return r
}
