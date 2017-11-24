package twitch

import(
    config   "github.com/codephobia/twitch-eos-thanks/server/config"
    database "github.com/codephobia/twitch-eos-thanks/server/database"
)

type Twitch struct {
    config   *config.Config
    database *database.Database
}

// create twitch
func NewTwitch(config *config.Config, database *database.Database) *Twitch {
    return &Twitch{
        config:   config,
        database: database,
    }
}

// get data from kraken
func (t *Twitch) Get() error {
    return nil
}