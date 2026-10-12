package config

type Proxy struct {
    Enable    bool   `json:"enable"`
    Host      string `json:"host"`
    Port      int    `json:"port"`
    TLSEnable bool   `json:"tls_enable"`
    SNI       string `json:"sni"`
}