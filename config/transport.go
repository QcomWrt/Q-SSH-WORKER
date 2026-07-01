package config

type Transport struct {
	TLS bool `json:"tls"`

	Host string `json:"host"`
	Path string `json:"path"`
	SNI  string `json:"sni"`
}