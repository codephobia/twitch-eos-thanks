package config

import (
    "encoding/json"
    "fmt"
    "os"
)

var (
    CONFIG_FILE string = "./config.json"
)

type Config struct {
    TwitchClientID   string `json:"twitch_client_id"`
    TwitchOAuthToken string `json:"twitch_oauth_token"`
    TwitchChannelID  string `json:"twitch_channel_id"`
    DBFileName       string `json:"db_file_name"`
    ApiHost          string `json:"api_host"`
    ApiPort          string `json:"api_port"`
    
    ClientTotalTime       int  `json:"client_total_time"`
    ClientShowFollowers   bool `json:"client_show_followers"`
    ClientShowSubscribers bool `json:"client_show_subscribers"`
}

// create new config
func NewConfig() *Config {
    return &Config{}
}

// load config file
func (c *Config) Load() error {
    configFile, err := os.Open(CONFIG_FILE)
    if err != nil {
        return fmt.Errorf("config open: %s", err)
    }
    defer configFile.Close()
    
    if err := json.NewDecoder(configFile).Decode(c); err != nil {
        return fmt.Errorf("config decode: %s", err)
    }
    
    return nil
}