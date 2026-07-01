package config

type Payload struct {
	Enable bool `json:"enable"`

	ProxyHost string `json:"proxy_host"`
	ProxyPort int    `json:"proxy_port"`

	Request string   `json:"request"`
	Expect  []string `json:"expect"`
}