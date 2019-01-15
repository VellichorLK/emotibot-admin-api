package UI

type Module struct {
	ID     int    `json:"id"`
	Code   string `json:"code"`
	URL    string `json:"url"`
	Enable bool   `json:"enable"`
}
