package Stats

import "time"

type AuditInput struct {
	Start       int          `json:"start_time"`
	End         int          `json:"end_time"`
	Filter      *AuditFilter `json:"filters"`
	Page        int          `json:"page"`
	ListPerPage int          `json:"limit"`
	Export		bool		 `json:"export"`
}

type AuditFilter struct {
	Module    string `json:"module"`
	Operation string `json:"operation"`
	UserID    string `json:"uid"`
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

type StatRow struct {
	UserQuery        string `json:"question"`
	Count            int    `json:"count"`
	StandardQuestion string `json:"std_q"`
	Score            int    `json:"score"`
	Answer           string `json:"answer"`
}

type StatRet struct {
	Data []*StatRow `json:"data"`
}

type DialogStatsRet struct {
    TableHeader    []DialogStatsHeader   `json:"table_header"`
    Data           []DialogStatsData	`json:"data"`
}

type DialogStatsHeader struct {
	Id    string    `json:"id"`
	Text  string	`json:"text"`
}

type DialogStatsData struct {
    Tag 		string		`json:"tag"`
    UserCnt 	int 		`json:"userCnt"`
    TotalCnt    int    		`json:"totalCnt"`
}


