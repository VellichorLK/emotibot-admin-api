package config

// Config is the struct describe a config in BFOP system
type Config struct {
	Code       string `json:"code"`
	Module     string `json:"module"`
	Value      string `json:"value"`
	UpdateTime int64  `json:"update_time"`
	Status     bool   `json:"status"`
}
