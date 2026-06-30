package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {

	if c.Listen.Port <= 0 {
		return fmt.Errorf("invalid listen port")
	}

	if c.SSH.Host == "" {
		return fmt.Errorf("ssh host required")
	}

	if c.SSH.Port <= 0 {
		return fmt.Errorf("invalid ssh port")
	}

	if c.SSH.Username == "" {
		return fmt.Errorf("ssh username required")
	}

	return nil
}