package Stats

import (
	"strings"
	"time"
)

type AuditInput struct {
	Start       int          `json:"start_time"`
	End         int          `json:"end_time"`
	Filter      *AuditFilter `json:"filters"`
	Page        int          `json:"page"`
	ListPerPage int          `json:"limit"`
	Export      bool         `json:"export"`
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
	TableHeader []DialogStatsHeader `json:"table_header"`
	Data        []DialogStatsData   `json:"data"`
}

type DialogStatsHeader struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type DialogStatsData struct {
	Tag      string `json:"tag"`
	UserCnt  int    `json:"userCnt"`
	TotalCnt int    `json:"totalCnt"`
}

// SessionCondition is used to create a conditioned query for sessions table
type SessionCondition struct {
	StartTime int64      `json:"start_time"`
	EndTime   int64      `json:"end_time"`
	Keyword   string     `json:"keyword"`
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Duration  int        `json:"duration"`
	Status    string     `json:"status"`
	Limit     *PageLimit `json:"limit"`
}

//sessionsStatStrToInt search string into database status
var SessionStatStrToInt = map[string]int{
	"canceled": -3,
	"timeout":  -2,
	"toHuman":  -1,
	"ongoing":  0,
	"finished": 1,
}

//sessionsStatIntToStr search int into represent text
var sessionStatIntToStr = map[int]string{
	-3: "canceled",
	-2: "timeout",
	-1: "toHuman",
	0:  "ongoing",
	1:  "finished",
}

//JoinedSQLCondition create a JOINED SQL condition based on SessionCondition & SessionTbl & recordsTbl
func (c *SessionCondition) JoinedSQLCondition(sessionTblName, recordTblName string) (preparedCond string, values []interface{}) {
	var sqlText string
	var AndConditions = []string{}
	values = []interface{}{}
	if c.StartTime != 0 || c.EndTime != 0 {
		query := sessionTblName + ".start_time >= ? AND " + sessionTblName + ".end_time <= ? "
		AndConditions = append(AndConditions, query)
		values = append(values, c.StartTime, c.EndTime)
	}
	if c.Keyword != "" {
		query := recordTblName + ".`user_q` LIKE ?"
		AndConditions = append(AndConditions, query)
		values = append(values, "%"+c.Keyword+"%")
	}
	if c.ID != "" {
		AndConditions = append(AndConditions, sessionTblName+".`id` = ?")
		values = append(values, c.ID)
	}
	if c.UserID != "" {
		AndConditions = append(AndConditions, recordTblName+".`user_id` = ?")
		values = append(values, c.UserID)
	}
	if c.Duration != 0 {
		query := sessionTblName + ".start_time - " + sessionTblName + ".end_time >= ?"
		AndConditions = append(AndConditions, query)
		values = append(values, c.Duration*1000)
	}

	//Remember to validate status at controller phase
	if c.Status != "" {
		query := sessionTblName + ".status = ?"
		AndConditions = append(AndConditions, query)
		status, _ := SessionStatStrToInt[c.Status]
		values = append(values, status)
	}
	sqlText = strings.Join(AndConditions, " AND ")
	return sqlText, values
}

// PageLimit is a sql query condition to limit which page and size to load
type PageLimit struct {
	Index    int `json:"index"`
	PageSize int `json:"page_size"`
}

// Session map to a sessions table row.
type Session struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	StartTime   int64       `json:"start_time"`
	EndTime     int64       `json:"end_time"`
	Duration    int64       `json:"duration"`
	Status      string      `json:"status"`
	Information []ValuePair `json:"information"`
	Notes       string      `json:"notes"`
}

//ValuePair represent a nlu's value data
type ValuePair struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
