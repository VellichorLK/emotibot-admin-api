package data

// App store basic app usage information

type AppV5 struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Status  int          `json:"status"`
	AppType int          `json:"app_type"` // app_type 0: Bot App  1: CCS App (中控)
	Props   []*AppPropV5 `json:"props"`
}

type AppDetailV5 struct {
	AppV5
	Description string `json:"description"`
}
