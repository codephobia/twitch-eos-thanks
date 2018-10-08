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

	config "github.com/codephobia/twitch-eos-thanks/server/config"
	database "github.com/codephobia/twitch-eos-thanks/server/database"
)

// API is the web api.
type API struct {
	config   *config.Config
	database *database.Database

	server *http.Server
}

// NewAPI returns a new api.
func NewAPI(c *config.Config, db *database.Database) *API {
	return &API{
		config:   c,
		database: db,
	}
}

// Init initializes the api.
func (api *API) Init() error {
	// create the server
	api.server = &http.Server{
		Handler:      handlers.CompressHandler(handlers.CORS()(api.Handler())),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// create a listener
	hostURL := strings.Join([]string{api.config.APIHost, ":", api.config.APIPort}, "")
	listener, err := net.Listen("tcp", hostURL)
	if err != nil {
		return fmt.Errorf("error starting api server: %s", err)
	}

	// run server
	log.Printf("API Server running: %s", listener.Addr().String())
	api.server.Serve(listener)

	return nil
}

// Handler handles incoming api routes.
func (api *API) Handler() http.Handler {
	// create router
	r := mux.NewRouter()

	// follow webhook
	r.Handle("/follow", api.handleFollow())

	// get followers
	r.Handle("/followers", api.handleFollowers())

	// get subscribers
	r.Handle("/subscribers", api.handleSubscribers())

	// get bits
	r.Handle("/bits", api.handleBits())

	// return router
	return r
}
