package config

type SSH struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`

	Username string `json:"username"`
	Password string `json:"password"`

	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}