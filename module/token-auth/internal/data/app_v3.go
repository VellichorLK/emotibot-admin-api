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

type BFAppV3 struct {
	ID   string `json:"appid"`
	Name string `json:"nickname"`
}

func (BFApp *BFAppV3) CopyFromApp(app *AppDetailV3) {
	BFApp.ID = app.ID
	BFApp.Name = app.Name
}
