package Stats

import "time"

type AuditInput struct {
	Start       int          `json:"start_time"`
	End         int          `json:"end_time"`
	Filter      *AuditFilter `json:"filters"`
	Page        int          `json:"page"`
	ListPerPage int          `json:"limit"`
}

type AuditFilter struct {
	Module    string `json:"module"`
	Operation string `json:"operation"`
	UserID    string `json:"user_id"`
}

type AuditLog struct {
	UserID     string    `json:"user_id"`
	Module     string    `json:"module"`
	Operation  string    `json:"operation"`
	Result     int       `json:"result"`
	CreateTime time.Time `json:"create_time"`
	UserIP     string    `json:"user_ip"`
	Content    string    `json:"content"`
}

type AuditRet struct {
	TotalCount int         `json:"total"`
	Data       []*AuditLog `json:"data"`
}
