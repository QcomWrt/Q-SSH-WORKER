package config

import (
	"encoding/json"
	"os"
)

func Load(file string) (*Config, error) {

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	setDefault(cfg)

	return cfg, nil
}