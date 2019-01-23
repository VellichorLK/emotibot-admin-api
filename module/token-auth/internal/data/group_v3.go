package data

type GroupV3 struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type GroupDetailV3 struct {
	GroupV3
	Apps []*AppV3 `json:"apps"`
}
