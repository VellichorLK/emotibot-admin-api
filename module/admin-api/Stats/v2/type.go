package v2

import "time"

type AuditInput struct {
	Page         int          `json:"page"`
	ListPerPage  int          `json:"limit"`
	Start        int          `json:"start_time"`
	End          int          `json:"end_time"`
	EnterpriseID string       `json:"enterprise_id"`
	RobotID      string       `json:"robot_id"`
	UserID       string       `json:"user_id"`
	Filter       *AuditFilter `json:"operation"`
	Export       bool
}

type AuditFilter struct {
	Module    string `json:"module"`
	Operation string `json:"type"`
}

type AuditLog struct {
	EnterpriseID string    `json:"enterprise"`
	AppID        string    `json:"appid"`
	UserID       string    `json:"user"`
	Module       string    `json:"module"`
	Operation    string    `json:"operation"`
	Result       int       `json:"result"`
	CreateTime   time.Time `json:"create_time"`
	UserIP       string    `json:"user_ip"`
	Content      string    `json:"content"`
}
