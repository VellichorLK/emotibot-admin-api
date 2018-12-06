package v1

import (
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
)

type Record struct {
	EnterpriseID string                 `json:"enterprise_id"`
	AppID        string                 `json:"app_id"`
	UserID       string                 `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	UserQ        string                 `json:"user_q"`
	StdQ         string                 `json:"std_q"`
	Answer       []data.Answer          `json:"answer"`
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

type VisitStatsQuery struct {
	data.CommonQuery
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
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        VisitStatsData         `json:"data"`
	Total       VisitStatsTotal        `json:"total"`
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
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        []VisitStatsTagData    `json:"data"`
	Total       VisitStatsQ            `json:"total"`
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
	TableHeader []data.TableHeaderItem   `json:"table_header"`
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

var VisitStatsTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "统计项",
		ID:   "time_txt",
	},
	data.TableHeaderItem{
		Text: "总会话数",
		ID:   common.VisitStatsMetricConversations,
	},
	data.TableHeaderItem{
		Text: "独立用户数",
		ID:   common.VisitStatsMetricUniqueUsers,
	},
	data.TableHeaderItem{
		Text: "新增用户数",
		ID:   common.VisitStatsMetricNewUsers,
	},
	data.TableHeaderItem{
		Text: "总提问数",
		ID:   common.VisitStatsMetricTotalAsks,
	},
	data.TableHeaderItem{
		Text: "标准回复",
		ID:   common.VisitStatsMetricNormalResponses,
	},
	data.TableHeaderItem{
		Text: "聊天",
		ID:   common.VisitStatsMetricChats,
	},
	data.TableHeaderItem{
		Text: "其他",
		ID:   common.VisitStatsMetricOthers,
	},
	data.TableHeaderItem{
		Text: "未知问题回复",
		ID:   common.VisitStatsMetricUnknownQnA,
	},
	data.TableHeaderItem{
		Text: "未解决",
		ID:   "unsolved",
	},
	data.TableHeaderItem{
		Text: "成功率",
		ID:   common.VisitStatsMetricSuccessRate,
	},
	data.TableHeaderItem{
		Text: "解决率",
		ID:   "solved_rate",
	},
	data.TableHeaderItem{
		Text: "平均会话对话数",
		ID:   common.VisitStatsMetricConversationsPerSession,
	},
}

var VisitStatsTagTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "总会话数",
		ID:   common.VisitStatsMetricConversations,
	},
	data.TableHeaderItem{
		Text: "独立用户数",
		ID:   common.VisitStatsMetricUniqueUsers,
	},
	data.TableHeaderItem{
		Text: "新增用户数",
		ID:   common.VisitStatsMetricNewUsers,
	},
	data.TableHeaderItem{
		Text: "总提问数",
		ID:   common.VisitStatsMetricTotalAsks,
	},
	data.TableHeaderItem{
		Text: "标准回复",
		ID:   common.VisitStatsMetricNormalResponses,
	},
	data.TableHeaderItem{
		Text: "聊天",
		ID:   common.VisitStatsMetricChats,
	},
	data.TableHeaderItem{
		Text: "其他",
		ID:   common.VisitStatsMetricOthers,
	},
	data.TableHeaderItem{
		Text: "未知问题回复",
		ID:   common.VisitStatsMetricUnknownQnA,
	},
	data.TableHeaderItem{
		Text: "未解决",
		ID:   "unsolved",
	},
	data.TableHeaderItem{
		Text: "成功率",
		ID:   common.VisitStatsMetricSuccessRate,
	},
	data.TableHeaderItem{
		Text: "解决率",
		ID:   "solved_rate",
	},
	data.TableHeaderItem{
		Text: "平均会话对话数",
		ID:   common.VisitStatsMetricConversationsPerSession,
	},
}

var AnswerCategoryTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "业务类",
		ID:   common.CategoryBusiness,
	},
	data.TableHeaderItem{
		Text: "聊天类",
		ID:   common.CategoryChat,
	},
	data.TableHeaderItem{
		Text: "其他",
		ID:   common.CategoryOther,
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
