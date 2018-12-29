package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	Gateway string `json:"gateway,omitempty"`
	Token   string `json:"token,omitempty"`
}

var defaultConfig = Config{}

func LoadConfig() (*Config, error) {
	var config Config
	if user, err := user.Current(); err != nil {
		return nil, err
	} else if _, err := os.Stat(filepath.Join(user.HomeDir, ".iot", "config.json")); os.IsNotExist(err) {
		config := defaultConfig
		return &config, nil
	} else if b, err := ioutil.ReadFile(filepath.Join(user.HomeDir, ".iot", "config.json")); err != nil {
		return nil, err
	} else if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) Save() error {
	if user, err := user.Current(); err != nil {
		return err
	} else if err := os.MkdirAll(filepath.Join(user.HomeDir, ".iot"), 0700); err != nil {
		return err
	} else if b, err := json.MarshalIndent(c, "", "  "); err != nil {
		return err
	} else if err := ioutil.WriteFile(filepath.Join(user.HomeDir, ".iot", "config.json"), b, 0600); err != nil {
		return err
	}
	return nil
}
