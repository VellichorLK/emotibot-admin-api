package clustering

import (
	"encoding/json"
)

//ReportStatus is a set of status code define for clustering report.
type ReportStatus int

//defined status code of clustering report
const (
	ReportStatusError     = -1
	ReportStatusRunning   = 0
	ReportStatusCompleted = 1
)

// Report represent a clustering task.
// It is a one to one mapping of the RDB table `reports`
type Report struct {
	ID          uint64
	CreatedTime int64
	UpdatedTime int64
	Condition   string
	UserID      string
	AppID       string
	//IgnoredSize represent counts of how many records status is ignored, which will not be included in the task.
	IgnoredSize int64
	//MarkedSize represent counts of how many records status is marked, which will not be included in the task too.
	MarkedSize int64
	//SkippedSize represent counts of how many records already is standard Question, which will not be included in the task.
	SkippedSize int64
	//Status default as 0 (running), 1(completed), -1 (error)
	Status ReportStatus
}

// ReportError represent error of a report.
// It is an one to one mapping for RDB table.
// BE CARE: **IT IS NOT A GOLANG ERROR SYNTAX!!**
type ReportError struct {
	ID         uint64
	ReportID   uint64
	Cause      string
	CreateTime int64
}

// Cluster is a subset of Report, contains userQuestions as a group
type Cluster struct {
	ID          uint64
	ReportID    uint64
	Tags        string
	CreatedTime int64
}

//ReportRecord is Report's record. an one to one mapping to RDB `report_records`
type ReportRecord struct {
	ID            uint64
	ReportID      uint64
	ClusterID     uint64
	ChatRecordID  string
	Content       string
	CreatedTime   int64
	IsCentralNode bool
}

type searchPeriod struct {
	StartTime *int64 `json:"start_time"`
	EndTime   *int64 `json:"end_time"`
}

//ReportQuery is a complex condition for querying reports
type ReportQuery struct {
	Reports     []uint64      `json:"reports"`
	CreatedTime *searchPeriod `json:"created_time"`
	UpdatedTime *searchPeriod `json:"updated_time"`
	UserID      *string       `json:"user_id"`
	Status      *int          `json:"status"`
	AppID       string
}

/*
SSMConfig is a highly dynamic struct that is very hard to parse.
All we need is the simpleft trained module. so we use interface{} with switching to minify struct fields.
*/
type ssmConfig struct {
	Items []ssmItem `json:"items"`
}

type ssmItem struct {
	Name string `json:"name"`
	//Value can contain multiple struct, we only want to parse as ssmValueElement
	Value *json.RawMessage `json:"value"`
}

type ssmValueElement struct {
	Name string `json:"name"`
	//Parameters can contains multiple struct, which we only want the ssmParameters struct
	Parameters *json.RawMessage `json:"parameters"`
}
type ssmParameters struct {
	Candidate string `json:"candidate,omitempty"`
	Data      string `json:"data,omitempty"`
	Model     string `json:"model,omitempty"`
}
