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
	Limit     int                        `json:"limit"`
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

//RecordQuery is the complete conditions for searching Records.
//All pointer variables including slice are optional conditions
//Any non-pointer variable SHOULD BE initialized properly.
type RecordQuery struct {
	Keyword   *string       `json:"keyword"`
	StartTime *int64        `json:"start_time"`
	EndTime   *int64        `json:"end_time"`
	Emotions  []string      `json:"emotions"`
	QTypes    []string      `json:"question_types"`
	Platforms []string      `json:"platforms"`
	Genders   []string      `json:"genders"`
	UserID    *string       `json:"uid"`
	Records   []interface{} `json:"records"`
	IsIgnored *bool         `json:"is_ignored"`
	IsMarked  *bool         `json:"is_marked"`
	From      int64
	Limit     int
	AppID     string
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

//VisitRecordsResponse is the schema of /record/query api response
type VisitRecordsResponse struct {
	Data        []*VisitRecordsData `json:"data"`
	Limit       int                 `json:"limit"`
	Page        int                 `json:"page"`
	MarkedSize  int64               `json:"marked_size"`
	IgnoredSize int64               `json:"ignored_size"`
	TableHeader []TableHeaderItem   `json:"table_header"`
	TotalSize   int64               `json:"total_size"`
}

type VisitRecordsData struct {
	UniqueID  string  `json:"id"`
	UserID    string  `json:"user_id"`
	UserQ     string  `json:"user_q"`
	Score     float64 `json:"score"`
	StdQ      string  `json:"std_q"`
	Answer    string  `json:"answer"`
	LogTime   string  `json:"log_time"`
	Emotion   string  `json:"emotion"`
	QType     string  `json:"qtype"`
	IsMarked  bool    `json:"is_marked"`
	IsIgnored bool    `json:"is_ignored"`
}

type VisitRecordsRawData struct {
	UniqueID  string   `json:"unique_id"`
	UserID    string   `json:"user_id"`
	UserQ     string   `json:"user_q"`
	Score     float64  `json:"score"`
	StdQ      string   `json:"std_q"`
	Answer    []Answer `json:"answer"`
	LogTime   string   `json:"log_time"`
	Emotion   string   `json:"emotion"`
	QType     string   `json:"qtype"`
	IsMarked  bool     `json:"isMarked"`
	IsIgnored bool     `json:"isIgnored"`
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
