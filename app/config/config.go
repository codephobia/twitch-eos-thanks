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
    TwitchClientID    string `json:"twitch_client_id"`
    TwitchOAuthToken  string `json:"twitch_oauth_token"`
    TwitchChannelID   string `json:"twitch_channel_id"`
    DBFileName        string `json:"db_file_name"`
    ApiHost           string `json:"api_host"`
    ApiPort           string `json:"api_port"`
    CodephobiaApiHost string `json:"codephobia_api_host"`
    CodephobiaApiPort string `json:"codephobia_api_port"`
    
    ClientTimeTotal         int  `json:"client_time_total"`
    ClientTimePer           int  `json:"client_time_per"`
    ClientShowFollowers     bool `json:"client_show_followers"`
    ClientShowSubscribers   bool `json:"client_show_subscribers"`
    ClientShowCurrentStream bool `json:"client_show_current_stream"`
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