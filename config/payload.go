package config

type Payload struct {
    Enable  bool     `json:"enable"`
    Request string   `json:"request"`
    Expect  []string `json:"expect"`
}