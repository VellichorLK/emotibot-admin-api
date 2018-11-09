package data

const VisitRecordsPageLimit = 20
const VisitRecordsExportPageLimit = 1000
const MaxNumRecordsPerXlsx    = 100000

const (
	XlsxExportDir           = "exports"
	XlsxDirTimestampFormat  = "20060102"
	XlsxFileTimestampFormat = "20060102_150405"
)

const (
	CodeExportTaskRunning = iota
	CodeExportTaskCompleted
	CodeExportTaskFailed
	CodeExportTaskEmpty
)

var ExportTaskCodesMap = map[int]string{
	CodeExportTaskRunning: "RUNNING",
	CodeExportTaskCompleted: "COMPLETED",
	CodeExportTaskFailed: "FAILED",
	CodeExportTaskEmpty: "EMPTY",
}

const (
	CategoryBusiness = "business"
	CategoryChat     = "chat"
	CategoryOther    = "other"
)

const (
	VisitRecordsMetricSessionID    = "session_id"
	VisitRecordsMetricAppID        = "app_id"
	VisitRecordsMetricUserID       = "user_id"
	VisitRecordsMetricUserQ        = "user_q"
	VisitRecordsMetricStdQ         = "std_q"
	VisitRecordsMetricAnswer       = "answer"
	VisitRecordsMetricModule       = "module"
	VisitRecordsMetricEmotion      = "emotion"
	VisitRecordsMetricEmotionScore = "emotion_score"
	VisitRecordsMetricIntent       = "intent"
	VisitRecordsMetricIntentScore  = "intent_score"
	VisitRecordsMetricLogTime      = "log_time"
	VisitRecordsMetricScore        = "score"
	VisitRecordsMetricCustomInfo   = "custom_info"
	VisitRecordsMetricNote         = "note"
	VisitRecordsMetricQType        = "qtype"
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

type VisitRecordsDataBase struct {
	SessionID string  `json:"session_id"`
	UserID    string  `json:"user_id"`
	UserQ     string  `json:"user_q"`
	Score     float64 `json:"score"`
	StdQ      string  `json:"std_q"`
	LogTime   string  `json:"log_time"`
	Emotion   string  `json:"emotion"`
	QType     string  `json:"qtype"`
	IsMarked  bool    `json:"is_marked"`
	IsIgnored bool    `json:"is_ignored"`
}

type VisitRecordsData struct {
	VisitRecordsDataBase
	UniqueID string `json:"id"`
	Answer   string `json:"answer"`
}

type VisitRecordsRawData struct {
	VisitRecordsDataBase
	UniqueID string   `json:"unique_id"`
	Answer   []Answer `json:"answer"`
}

type VisitRecordsHitResult struct {
	VisitRecordsRawData
	Module string `json:"module"`
}

type VisitRecordsExportBase struct {
	SessionID    string  `json:"session_id"`
	AppID        string  `json:"app_id"`
	UserID       string  `json:"user_id"`
	UserQ        string  `json:"user_q"`
	StdQ         string  `json:"std_q"`
	Module       string  `json:"module"`
	Emotion      string  `json:"emotion"`
	EmotionScore float64 `json:"emotion_score"`
	Intent       string  `json:"intent"`
	IntentScore  float64 `json:"intent_score"`
	LogTime      string  `json:"log_time"`
	Score        float64 `json:"score"`
	Note         string  `json:"note"`
}

type VisitRecordsExportData struct {
	VisitRecordsExportBase
	Answer     string `json:"answer"`
	CustomInfo string `json:"custom_info"`
}

type VisitRecordsExportRawData struct {
	VisitRecordsExportBase
	Answer     []Answer               `json:"answer"`
	CustomInfo map[string]interface{} `json:"custom_info"`
}

var VisitRecordsTableHeader = []TableHeaderItem{
	TableHeaderItem{
		Text: "会话ID",
		ID:   VisitRecordsMetricSessionID,
	},
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
	"会话ID",
	"用户ID",
	"用户问题",
	"标准问题",
	"机器人回答",
	"匹配分数",
	"出话模组",
	"访问时间",
	"情感",
	"情感分数",
	"意图",
	"意图分数",
	"客制化资讯",
	"附注",
}

type VisitRecordsExportResponse struct {
	ExportID string `json:"export_id"`
}

type VisitRecordsExportStatusResponse struct {
	Status string `json:"status"`
}
