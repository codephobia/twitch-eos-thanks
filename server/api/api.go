package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "strings"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
    
    config   "github.com/codephobia/twitch-eos-thanks/server/config"
    database "github.com/codephobia/twitch-eos-thanks/server/database"
)

// api struct
type Api struct {
    config   *config.Config
    database *database.Database
    
    server   *http.Server
}

// create new api
func NewApi(c *config.Config, db *database.Database) *Api {
    return &Api{
        config:   c,
        database: db,
    }
}

// init api
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
    
    // follow webhook
    r.Handle("/follow", api.handleFollow())

    // get followers
    r.Handle("/followers", api.handleFollowers())

    // return router
    return r
}

type ErrorResp struct {
    err string `json:"error"`
}

func (api *Api) handleError(w http.ResponseWriter, status int, err error) {
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(status)
    
    enc := json.NewEncoder(w)
    enc.Encode(&ErrorResp{
        err: err.Error(),
    })
}