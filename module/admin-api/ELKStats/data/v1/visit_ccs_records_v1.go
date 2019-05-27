package v1

type VisitCcsRecordsData struct {
	VisitRecordsDataBase
	UniqueID         string     `json:"unique_id"`
	CaseId			 string 	`json:"case_id"`
	Dataset			 string 	`json:"dataset"`
	TaskId			 string 	`json:"task_id"`
	AppId			 string						`json:"app_id"`
	RawResponse		 map[string]interface{} 	`json:"raw_reponse"`
	RuleIds          []string     				`json:"rule_ids"`
	DivertTags       []string     				`json:"divert_tags"`
	RankTags         []string     				`json:"rank_tags"`
}

type VisitCcsRecordsResponse struct {
	Data        []*VisitCcsRecordsData    `json:"data"`
	Limit       int                    `json:"limit"`
	Page        int                    `json:"page"`
	TotalSize   int64                  `json:"total_size"`
}

type CcsRecordQuery struct {
	Keyword      *string       `json:"keyword,omitempty"`
	StartTime    *int64        `json:"start_time,omitempty"`
	EndTime      *int64        `json:"end_time,omitempty"`
	Emotions     []string      `json:"emotions,omitempty"`
	QTypes       []string      `json:"question_types,omitempty"`
	Platforms    []string      `json:"platforms,omitempty"`
	Genders      []string      `json:"genders,omitempty"`
	UserID       *string       `json:"uid,omitempty"`
	Records      []interface{} `json:"records,omitempty"`
	IsIgnored    *bool         `json:"is_ignored,omitempty"`
	IsMarked     *bool         `json:"is_marked,omitempty"`
	From         int64         `json:"-"`
	Limit        int           `json:"-"`
	EnterpriseID string        `json:"-"`
	AppID        string        `json:"-"`
	TaskID		 string		   `json:"-"`
}