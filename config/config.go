package config

type Config struct {
	Listen    ListenConfig    `json:"listen"`
	SSH       SSHConfig       `json:"ssh"`
	Transport TransportConfig `json:"transport"`
	Payload   PayloadConfig   `json:"payload"`
}