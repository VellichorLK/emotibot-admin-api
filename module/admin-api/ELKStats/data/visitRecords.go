package data

const VisitRecordsPageLimit = 20
const VisitRecordsExportPageLimit = 1000

const (
	CategoryBusiness = "business"
	CategoryChat     = "chat"
	CategoryOther    = "other"
)

const (
	VisitRecordsMetricUserID  = "user_id"
	VisitRecordsMetricUserQ   = "user_q"
	VisitRecordsMetricScore   = "score"
	VisitRecordsMetricStdQ    = "std_q"
	VisitRecordsMetricAnswer  = "answer"
	VisitRecordsMetricLogTime = "log_time"
	VisitRecordsMetricEmotion = "emotion"
	VisitRecordsMetricQType   = "qtype"
)

type VisitRecordsTag struct {
	Type string
	Text string
}

type VisitRecordsRequest struct {
	StartTime int64                      `json:"start_time"`
	EndTime   int64                      `json:"end_time"`
	Filter    *VisitRecordsRequestFilter `json:"filter"`
	Page      int                        `json:"page"`
	Export    bool                       `json:"export"`
}

type VisitRecordsRequestFilter struct {
	Contains *VisitRecordsRequestFilterContains `json:"contains"`
	Emotions []VisitRecordsRequestFilterEmotion `json:"emotions"`
	QTypes   []VisitRecordsRequestFilterQType   `json:"qtypes"`
	Tags     []VisitRecordsRequestFilterTag     `json:"categories"`
	UserID   string                             `json:"uid"`
}

type VisitRecordsRequestFilterContains struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type VisitRecordsRequestFilterEmotion struct {
	Type  string                                  `json:"type"`
	Group []VisitRecordsRequestFilterEmotionGroup `json:"group"`
}

type VisitRecordsRequestFilterEmotionGroup struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type VisitRecordsRequestFilterQType struct {
	Type  string                                `json:"type"`
	Group []VisitRecordsRequestFilterQTypeGroup `json:"group"`
}

type VisitRecordsRequestFilterQTypeGroup struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type VisitRecordsRequestFilterTag struct {
	Type  string                              `json:"type"`
	Group []VisitRecordsRequestFilterTagGroup `json:"group"`
}

type VisitRecordsRequestFilterTagGroup struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type VisitRecordsQuery struct {
	CommonQuery
	Question  string
	UserID    string
	Emotions  []interface{}
	QType     string
	Tags      []VisitRecordsQueryTag
	Page      int
	PageLimit int
}

type VisitRecordsQueryTag struct {
	Type  string
	Texts []string
}

type VisitRecordsResponse struct {
	Data        []*VisitRecordsData `json:"data"`
	Limit       int                 `json:"limit"`
	Page        int                 `json:"page"`
	TableHeader []TableHeaderItem   `json:"table_header"`
	TotalSize   int64               `json:"total_size"`
}

type VisitRecordsData struct {
	UserID  string  `json:"user_id"`
	UserQ   string  `json:"user_q"`
	Score   float64 `json:"score"`
	StdQ    string  `json:"std_q"`
	Answer  string  `json:"answer"`
	LogTime string  `json:"log_time"`
	Emotion string  `json:"emotion"`
	QType   string  `json:"qtype"`
}

type VisitRecordsRawData struct {
	UserID  string   `json:"user_id"`
	UserQ   string   `json:"user_q"`
	Score   float64  `json:"score"`
	StdQ    string   `json:"std_q"`
	Answer  []Answer `json:"answer"`
	LogTime string   `json:"log_time"`
	Emotion string   `json:"emotion"`
	QType   string   `json:"qtype"`
}

type VisitRecordsHitResult struct {
	VisitRecordsRawData
	Module string `json:"module"`
}

var VisitRecordsTableHeader = []TableHeaderItem{
	TableHeaderItem{
		Text: "用户ID",
		ID:   VisitRecordsMetricUserID,
	},
	TableHeaderItem{
		Text: "用户问题",
		ID:   VisitRecordsMetricUserQ,
	},
	TableHeaderItem{
		Text: "匹配分数",
		ID:   VisitRecordsMetricScore,
	},
	TableHeaderItem{
		Text: "标准问题",
		ID:   VisitRecordsMetricStdQ,
	},
	TableHeaderItem{
		Text: "机器人回答",
		ID:   VisitRecordsMetricAnswer,
	},
	TableHeaderItem{
		Text: "访问时间",
		ID:   VisitRecordsMetricLogTime,
	},
	TableHeaderItem{
		Text: "情感",
		ID:   VisitRecordsMetricEmotion,
	},
	TableHeaderItem{
		Text: "问答类别",
		ID:   VisitRecordsMetricQType,
	},
}

var VisitRecordsExportHeader = []string{
	"用户ID",
	"用户问题",
	"匹配分数",
	"标准问题",
	"机器人回答",
	"访问时间",
	"情感",
	"问答类别",
}
