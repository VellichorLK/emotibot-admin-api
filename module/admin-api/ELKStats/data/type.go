package data

import (
	"time"
)

type CommonQuery struct {
	EnterpriseID string
	AppID        string
	StartTime    time.Time
	EndTime      time.Time
}

type Tag struct {
	Code string
	Name string
}

type TableHeaderItem struct {
	Text string `json:"text"`
	ID   string `json:"id"`
}
