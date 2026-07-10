package config

type Config struct {
    Listen    Listen    `json:"listen"`
    SSH       SSH       `json:"ssh"`
    Network   Network   `json:"network"`
    Proxy     Proxy     `json:"proxy"`
    Transport Transport `json:"transport"`
    Payload   Payload   `json:"payload"`
    // Mengikat struct Worker (dari worker.go) ke tag JSON "concurrency"
    Worker    Worker    `json:"concurrency"` 
}