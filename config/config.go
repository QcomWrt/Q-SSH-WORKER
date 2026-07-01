package config

type Config struct {
	Listen    Listen    `json:"listen"`
	SSH       SSH       `json:"ssh"`
	Network   Network   `json:"network"`
	Transport Transport `json:"transport"`
	Payload   Payload   `json:"payload"`
	Worker    Worker    `json:"worker"`
}