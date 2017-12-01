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
    MongoDBHost     string `json:"mongo_db_host"`
    MongoDBPort     string `json:"mongo_db_port"`
    MongoDBDatabase string `json:"mongo_db_database"`
    
    ApiHost string `json:"api_host"`
    ApiPort string `json:"api_port"`
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