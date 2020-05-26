package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Database DatabaseConfig

	Import ImportConfig

	Orbs OrbsClientConfig
}

func GetConfig(path string) (*Config, error) {
	rawJson, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(rawJson, config)
	return config, err
}
