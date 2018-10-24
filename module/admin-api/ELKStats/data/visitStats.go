package data

import (
	"strconv"
	"time"
)

const (
	VisitStatsMetricConversations           = "conversations"
	VisitStatsMetricUniqueUsers             = "unique_users"
	VisitStatsMetricNewUsers                = "new_users"
	VisitStatsMetricTotalAsks               = "total_asks"
	VisitStatsMetricNormalResponses         = "normal_responses"
	VisitStatsMetricChats                   = "chats"
	VisitStatsMetricOthers                  = "others"
	VisitStatsMetricUnknownQnA              = "unknown_qna"
	VisitStatsMetricSuccessRate             = "success_rate"
	VisitStatsMetricConversationsPerSession = "conversation_per_session"
)

const (
	VisitStatsTypeTime     = "time"
	VisitStatsTypeBarchart = "barchart"
)

const (
	VisitStatsFilterCategory = "category"
	VisitStatsFilterQType    = "qtype"
)

const (
	VisitQuestionsTypeTop    = "top"
	VisitQuestionsTypeUnused = "unused"
)

type Record struct {
	EnterpriseID string                 `json:"enterprise_id"`
	AppID        string                 `json:"app_id"`
	UserID       string                 `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	UserQ        string                 `json:"user_q"`
	StdQ         string                 `json:"std_q"`
	Answer       []Answer               `json:"answer"`
	Module       string                 `json:"module"`
	Emotion      string                 `json:"emotion"`
	EmotionScore float64                `json:"emotion_score"`
	Intent       string                 `json:"intent"`
	IntentScore  float64                `json:"intent_score"`
	LogTime      string                 `json:"log_time"`
	Score        float64                `json:"score"`
	CustomInfo   map[string]interface{} `json:"custom_info"`
	Host         string                 `json:"host"`
	UniqueID     string                 `json:"unique_id"`
	Note         string                 `json:"note"`
}

type Answer struct {
	Type       string        `json:"type"`
	SubType    string        `json:"subType"`
	Value      string        `json:"value"`
	Data       []interface{} `json:"data"`
	ExtendData string        `json:"extendData"`
}

type VisitStatsQuery struct {
	CommonQuery
	AggBy       string
	AggInterval string
	AggTagType  string
}

type Question struct {
	Question string
	Count    int64
}

type Questions []Question

func (q Questions) Len() int {
	return len(q)
}

func (q Questions) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q Questions) Less(i, j int) bool {
	return q[i].Count < q[j].Count
}

type UnmatchQuestion struct {
	Question   string
	Count      int64
	MaxLogTime string
	MinLogTime string
}

type VisitStatsResponse struct {
	TableHeader []TableHeaderItem `json:"table_header"`
	Data        VisitStatsData    `json:"data"`
	Total       VisitStatsTotal   `json:"total"`
}

type VisitStatsData struct {
	VisitStatsQuantities []VisitStatsQuantity `json:"quantities"`
	Type                 string               `json:"type"`
	Name                 string               `json:"name"`
}

type VisitStatsTotal struct {
	VisitStatsQ
	TimeText string `json:"time_txt"`
	Time     string `json:"time,omitempty"`
}

type VisitStatsTagResponse struct {
	TableHeader []TableHeaderItem   `json:"table_header"`
	Data        []VisitStatsTagData `json:"data"`
	Total       VisitStatsQ         `json:"total"`
}

type VisitStatsTagData struct {
	Q    VisitStatsQ `json:"q"`
	ID   string      `json:"id"`
	Name string      `json:"name"`
}

type VisitStatsQuantity struct {
	VisitStatsQ
	TimeText string `json:"time_txt"`
	Time     string `json:"time"`
}

type VisitStatsQuantities []VisitStatsQuantity

func (q VisitStatsQuantities) Len() int {
	return len(q)
}

func (q VisitStatsQuantities) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q VisitStatsQuantities) Less(i, j int) bool {
	qTimeI, err := strconv.ParseInt(q[i].Time, 10, 64)
	if err != nil {
		return false
	}

	_qTimeI := time.Unix(qTimeI, 0)

	qTimeJ, err := strconv.ParseInt(q[j].Time, 10, 64)
	if err != nil {
		return false
	}

	_qTimeJ := time.Unix(qTimeJ, 0)

	return _qTimeI.Before(_qTimeJ)
}

type AnswerCategoryStatsResponse struct {
	TableHeader []TableHeaderItem        `json:"table_header"`
	Data        []AnswerCategoryStatData `json:"data"`
	Total       VisitStatsQ              `json:"total"`
}

type AnswerCategoryStatData struct {
	Q    VisitStatsQ `json:"q"`
	ID   string      `json:"id"`
	Name string      `json:"name"`
}

func NewAnswerCategoryStatData(id string, name string) *AnswerCategoryStatData {
	return &AnswerCategoryStatData{
		Q:    *NewVisitStatsQ(),
		ID:   id,
		Name: name,
	}
}

type TopQuestionsResponse struct {
	Data []TopQuestionData `json:"data"`
}

type TopQuestionData struct {
	Q        int64  `json:"q"`
	Path     string `json:"path"`
	Question string `json:"question"`
	Rank     int    `json:"rank"`
}

type TopUnmatchedQuestionsResponse struct {
	Data []TopUnmatchedQuestionData `json:"data"`
}

type TopUnmatchedQuestionData struct {
	Question      string `json:"question"`
	Rank          int    `json:"rank"`
	Q             int64  `json:"q"`
	FirstTime     string `json:"first_time"`
	FirstTimeText string `json:"first_time_txt"`
	LastTime      string `json:"last_time"`
	LastTimeText  string `json:"last_time_txt"`
}

var VisitStatsTableHeader = []TableHeaderItem{
	TableHeaderItem{
		Text: "统计项",
		ID:   "time_txt",
	},
	TableHeaderItem{
		Text: "总会话数",
		ID:   VisitStatsMetricConversations,
	},
	TableHeaderItem{
		Text: "独立用户数",
		ID:   VisitStatsMetricUniqueUsers,
	},
	TableHeaderItem{
		Text: "新增用户数",
		ID:   VisitStatsMetricNewUsers,
	},
	TableHeaderItem{
		Text: "总提问数",
		ID:   VisitStatsMetricTotalAsks,
	},
	TableHeaderItem{
		Text: "标准回复",
		ID:   VisitStatsMetricNormalResponses,
	},
	TableHeaderItem{
		Text: "聊天",
		ID:   VisitStatsMetricChats,
	},
	TableHeaderItem{
		Text: "其他",
		ID:   VisitStatsMetricOthers,
	},
	TableHeaderItem{
		Text: "未知问题回复",
		ID:   VisitStatsMetricUnknownQnA,
	},
	TableHeaderItem{
		Text: "未解决",
		ID:   "unsolved",
	},
	TableHeaderItem{
		Text: "成功率",
		ID:   VisitStatsMetricSuccessRate,
	},
	TableHeaderItem{
		Text: "解决率",
		ID:   "solved_rate",
	},
	TableHeaderItem{
		Text: "平均会话对话数",
		ID:   VisitStatsMetricConversationsPerSession,
	},
}

var AnswerCategoryTableHeader = []TableHeaderItem{
	TableHeaderItem{
		Text: "业务类",
		ID:   CategoryBusiness,
	},
	TableHeaderItem{
		Text: "聊天类",
		ID:   CategoryChat,
	},
	TableHeaderItem{
		Text: "其他",
		ID:   CategoryOther,
	},
}

type VisitStatsQ struct {
	Conversations          int64  `json:"conversations"`
	UniqueUsers            int64  `json:"unique_users"`
	NewUsers               int64  `json:"new_users"`
	TotalAsks              int64  `json:"total_asks"`
	NormalResponses        int64  `json:"normal_responses"`
	Chats                  int64  `json:"chats"`
	Others                 int64  `json:"others"`
	UnknownQnA             int64  `json:"unknown_qna"`
	Unsolved               int64  `json:"unsolved"`
	SuccessRate            string `json:"success_rate"`
	SolvedRate             string `json:"solved_rate"`
	ConversationPerSession string `json:"conversation_per_session"`
	Solved                 int64  `json:"solved"`
}

func NewVisitStatsQ() *VisitStatsQ {
	return &VisitStatsQ{
		Conversations:          0,
		UniqueUsers:            0,
		NewUsers:               0,
		TotalAsks:              0,
		NormalResponses:        0,
		Chats:                  0,
		Others:                 0,
		UnknownQnA:             0,
		Unsolved:               0,
		SuccessRate:            "N/A",
		SolvedRate:             "N/A",
		ConversationPerSession: "N/A",
		Solved:                 0,
	}
}

type VisitStatsQueryHandler func(query VisitStatsQuery) (map[string]interface{}, error)
