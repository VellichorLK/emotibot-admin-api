package v2

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
)

type VisitCcsRecordsRequest struct {
	StartTime     int64    `json:"start_time"`
	EndTime       int64    `json:"end_time"`
	Page          *int64   `json:"page"`
	Limit         *int64   `json:"limit"`
	SessionID     *string  `json:"session_id,omitempty"`
}

type VisitCcsRecordsQuery struct {
	data.CommonQuery
	From          int64
	Limit         int64
	SessionID     *string
}

type VisitCcsRecordsResponse struct {
	TableHeader []data.TableHeaderItem 		`json:"table_header"`
	Data        []*VisitCcsRecordsData 		`json:"data"`
	Limit       int64                  		`json:"limit"`
	Page        int64                  		`json:"page"`
	TotalSize   int64                  		`json:"total_size"`
}

type VisitCcsRecordsData struct {
	SessionID   string  	`json:"session_id"`
	UserID      string  	`json:"user_id"`
	UserQ       string  	`json:"user_q"`
	LogTime     string  	`json:"log_time"`
	RawResponse string 		`json:"raw_response"`
	Answers		[]string 	`json:"answers"`
	AIModule	string		`json:"ai_module"`
}

type VisitCcsRecordsRawData struct {
	SessionID   string  	`json:"session_id"`
	UserID      string  	`json:"user_id"`
	UserQ       string  	`json:"user_q"`
	LogTime     string  	`json:"log_time"`
	RawResponse string 		`json:"raw_response"`
	Answers		[]string 	`json:"answers"`
	AIModule	string		`json:"ai_module"`
}

type VisitCcsRecordsQueryResult struct {
	Data        []*VisitCcsRecordsData
	TotalSize   int64
}

var VisitCcsRecordsTableHeader = []data.TableHeaderItem{
	{
		Text: "会话ID",
		ID:   common.VisitRecordsMetricSessionID,
	},
	{
		Text: "用户ID",
		ID:   common.VisitRecordsMetricUserID,
	},
	{
		Text: "用户问题",
		ID:   common.VisitRecordsMetricUserQ,
	},
	{
		Text: "访问时间",
		ID:   common.VisitRecordsMetricLogTime,
	},
	{
		Text: "回答",
		ID:   common.VisitTagsMetricAnswers,
	},
	{
		Text: "AI模块",
		ID:   common.VisitTagsMetricAIModule,
	},
}
