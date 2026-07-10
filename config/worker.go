package config

type Worker struct {
	// --- Fitur Multi-Worker Baru ---
	Enable    bool `json:"enable"`
	Workers   int  `json:"workers"`
	StartPort int  `json:"start_port"`

	// --- Fitur Lama Kamu (Tetap Dipertahankan) ---
	MaxRetry    int  `json:"max_retry"`
	RetryDelay  int  `json:"retry_delay"`
	HealthCheck bool `json:"health_check"`
}