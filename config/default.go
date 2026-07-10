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

	// ======================================================================
	// 🟢 LOGIKA FALLBACK DEFAULT UNTUK MULTI-WORKER / CONCURRENCY
	// ======================================================================
	
	// Jika user tidak menentukan port awal multi-worker, samakan dengan port default listen (1080)
	if cfg.Worker.StartPort == 0 {
		cfg.Worker.StartPort = cfg.Listen.Port
	}

	// Jika max_retry kosong di JSON, kunci ke angka aman 3 kali percobaan dial
	if cfg.Worker.MaxRetry == 0 {
		cfg.Worker.MaxRetry = 3
	}

	// Jika retry_delay kosong di JSON, berikan jeda aman 5 detik antar dial balik
	if cfg.Worker.RetryDelay == 0 {
		cfg.Worker.RetryDelay = 5
	}

	// Jika multi-worker aktif, otomatis paksa HealthCheck bernilai true demi kestabilan pasukannya
	if cfg.Worker.Enable && !cfg.Worker.HealthCheck {
		cfg.Worker.HealthCheck = true
	}
}