package data

type Session struct {
	ID        string `json:"session_id"`
	Status    int    `json:"status"`
	Data      string `json:"data"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}
