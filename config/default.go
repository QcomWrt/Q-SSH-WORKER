package config

func setDefault(cfg *Config) {

	if cfg.Listen.Host == "" {
		cfg.Listen.Host = "127.0.0.1"
	}

	if cfg.Listen.Port == 0 {
		cfg.Listen.Port = 1080
	}

	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = 22
	}

	if cfg.Network.Type == "" {
		cfg.Network.Type = "tcp"
	}

	if cfg.Transport.Path == "" {
		cfg.Transport.Path = "/"
	}
}