package data

// App store basic app usage information
type AppV4 struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
	// app_type 0: Bot App  1: CCS App (中控)
	AppType int `json:"app_type"`
}

type AppDetailV4 struct {
	AppV4
	Description string `json:"description"`
}
