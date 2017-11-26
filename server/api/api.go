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
    twitch   "github.com/codephobia/twitch-eos-thanks/server/twitch"
)

type Api struct {
    config   *config.Config
    database *database.Database
    twitch   *twitch.Twitch
    
    server   *http.Server
}

func NewApi(config *config.Config, database *database.Database, twitch *twitch.Twitch) *Api {
    return &Api{
        config: config,
        database: database,
        twitch: twitch,
    }
}

func (api *Api) Init() error {
    // create the server
    api.server = &http.Server{
        Handler:        handlers.CompressHandler(handlers.CORS()(api.Handler())),
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
    }
    
    // create a listener
    hostUrl := strings.Join([]string{api.config.ApiHost, ":", api.config.ApiPort}, "")
    ln, err := net.Listen("tcp", hostUrl)
    if err != nil {
        return fmt.Errorf("error starting api server: %s", err)
    }
    
    // run server
    log.Printf("API Server running: %s", ln.Addr().String())
    api.server.Serve(ln)

    return nil
}

func (api *Api) Handler() http.Handler {
    // create router
    r := mux.NewRouter()
    
    // followers
    r.Handle("/followers", api.handleFollowers())
    
    return r
}

func (api *Api) handleFollowers() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
            case "GET":
                api.handleFollowersGet(w, r)
            default:
                api.handleError(w, 400, fmt.Errorf("method not allowed"))
        }
    })
}

// handleFollowersGet
func (api *Api) handleFollowersGet(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    enc := json.NewEncoder(w)
    enc.Encode(api.twitch.Followers)
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