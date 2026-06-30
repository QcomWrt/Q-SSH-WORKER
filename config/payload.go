package config

type TransportConfig struct {
	Type string `json:"type"`

	TLS  bool   `json:"tls"`
	SNI  string `json:"sni"`

	Host string `json:"host"`
	Path string `json:"path"`
}

type PayloadConfig struct {
	Enable    bool     `json:"enable"`
	ProxyHost string   `json:"proxy_host"`
	ProxyPort int      `json:"proxy_port"`
	Request   string   `json:"request"`
	Expect    []string `json:"expect"`
}