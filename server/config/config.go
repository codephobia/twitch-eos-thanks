package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	configFilePath = "./config.json"
)

// Config stores the configuration file options.
type Config struct {
	TwitchClientID            string `json:"twitch_client_id"`
	TwitchClientSecret        string `json:"twitch_client_secret"`
	TwitchOAuthToken          string `json:"twitch_oauth_token"`
	TwitchChannelID           string `json:"twitch_channel_id"`
	TwitchChannelOAuthToken   string `json:"twitch_channel_oauth_token"`
	TwitchChannelRefreshToken string `json:"twitch_channel_refresh_token"`

	MongoDBHost     string `json:"mongo_db_host"`
	MongoDBPort     string `json:"mongo_db_port"`
	MongoDBDatabase string `json:"mongo_db_database"`

	APIHost string `json:"api_host"`
	APIPort string `json:"api_port"`
}

// NewConfig returns a new config.
func NewConfig() *Config {
	return &Config{}
}

// Load loads the configuration file.
func (c *Config) Load() error {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("config open: %s", err)
	}
	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(c); err != nil {
		return fmt.Errorf("config decode: %s", err)
	}

	return nil
}

// Save saves the current in memory config values to
// the configuration json file.
func (c *Config) Save() error {
	// marshal config
	configJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// make json pretty
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, configJSON, "", "    "); err != nil {
		return err
	}

	// write updates to file
	return ioutil.WriteFile(configFilePath, prettyJSON.Bytes(), 0644)
}
