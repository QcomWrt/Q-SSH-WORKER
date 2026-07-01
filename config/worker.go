package config

type Worker struct {
	MaxRetry int `json:"max_retry"`

	RetryDelay int `json:"retry_delay"`

	HealthCheck bool `json:"health_check"`
}