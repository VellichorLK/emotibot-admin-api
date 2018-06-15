package data

// App store basic app usage information in it
type AppV3 struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type AppDetailV3 struct {
	AppV3
	Description string `json:"description"`
}
