package v2

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
)

type VisitRecordsRequest struct {
	StartTime     int64    `json:"start_time"`
	EndTime       int64    `json:"end_time"`
	Modules       []string `json:"modules,omitempty"`
	Platforms     []string `json:"platforms,omitempty"`
	Genders       []string `json:"genders,omitempty"`
	Emotions      []string `json:"emotions,omitempty"`
	IsIgnored     *bool    `json:"is_ignored,omitempty"`
	IsMarked      *bool    `json:"is_marked,omitempty"`
	MarkedIntent  *int64   `json:"marked_intent,omitempty"`
	Keyword       *string  `json:"keyword,omitempty"`
	UserID        *string  `json:"uid,omitempty"`
	SessionID     *string  `json:"session_id,omitempty"`
	TESessionID   *string  `json:"taskengine_session_id,omitempty"`
	Intent        *string  `json:"intent"`
	MinScore      *int64   `json:"min_score,omitempty"`
	MaxScore      *int64   `json:"max_score,omitempty"`
	LowConfidence *int64   `json:"low_confidence_score,omitempty"`
	FaqCategories []int64  `json:"faq_cats,omitempty"`
	FaqRobotTags  []int64  `json:"faq_robot_tags,omitempty"`
	Feedback      *string  `json:"feedback,omitempty"`
	Page          *int64   `json:"page"`
	Limit         *int64   `json:"limit"`
}

type VisitRecordsQuery struct {
	data.CommonQuery
	Modules       []string
	Platforms     []string
	Genders       []string
	Emotions      []string
	IsIgnored     *bool
	IsMarked      *bool
	Keyword       *string
	UserID        *string
	SessionID     *string
	TESessionID   *string
	Intent        *string
	MinScore      *int64
	MaxScore      *int64
	LowConfidence *int64
	FaqCategories []int64
	FaqRobotTags  []int64
	Feedback      *string
	From          int64
	Limit         int64
	RecordIDs     []interface{}
}

type VisitRecordsResponse struct {
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        []*VisitRecordsData    `json:"data"`
	IgnoredSize int64                  `json:"ignored_size"`
	MarkedSize  int64                  `json:"marked_size"`
	Limit       int64                  `json:"limit"`
	Page        int64                  `json:"page"`
	TotalSize   int64                  `json:"total_size"`
}

type VisitRecordsDataBase struct {
	SessionID    string  `json:"session_id"`
	TESessionID  string  `json:"taskengine_session_id"`
	UserID       string  `json:"user_id"`
	UserQ        string  `json:"user_q"`
	Score        float64 `json:"score"`
	StdQ         string  `json:"std_q"`
	LogTime      string  `json:"log_time"`
	Emotion      string  `json:"emotion"`
	EmotionScore float64 `json:"emotion_score"`
	Intent       string  `json:"intent"`
	IntentScore  float64 `json:"intent_score"`
	Module       string  `json:"module"`
	Source       string  `json:"source"`
}

type VisitRecordsRawData struct {
	VisitRecordsDataBase
	EmotionScore   float64                `json:"emotion_score"`
	IntentScore    float64                `json:"intent_score"`
	UniqueID       string                 `json:"unique_id"`
	Answer         []data.Answer          `json:"answer"`
	RawAnswer      string                 `json:"raw_answer"`
	CustomInfo     map[string]interface{} `json:"custom_info"`
	Note           string                 `json:"note"`
	IsMarked       bool                   `json:"isMarked"`
	MarkedIntent   *int64                 `json:"marked_intent"`
	IsIgnored      bool                   `json:"isIgnored"`
	FaqCategoryID  int64                  `json:"faq_cat_id"`
	FaqRobotTagIDs []int64                `json:"faq_robot_tag_id"`
	Feedback       string                 `json:"feedback"`
	CustomFeedback string                 `json:"custom_feedback"`
	FeedbackTime   int64                  `json:"feedback_time"`
	Threshold      int64                  `json:"threshold"`
	TSpan          int64                  `json:"tspan"`
}

type VisitRecordsCommon struct {
	VisitRecordsDataBase
	Answer           string  `json:"answer"`
	RawAnswer        string  `json:"raw_answer"`
	FaqCategoryName  string  `json:"faq_cat_name"`
	FaqRobotTagNames string  `json:"faq_robot_tag_name"`
	FaqCategoryID    int64   `json:"-"` // Redundant field for FAQ category name conversion
	FaqRobotTagIDs   []int64 `json:"-"` // Redundant field for FAQ robot tag name conversion
	Feedback         string  `json:"feedback"`
	CustomFeedback   string  `json:"custom_feedback"`
	FeedbackTime     string  `json:"feedback_time"`
	Threshold        int64   `json:"threshold"`
	TSpan            int64   `json:"tspan"`
}
type VisitRecordsData struct {
	VisitRecordsCommon
	UniqueID     string `json:"id"`
	IsMarked     bool   `json:"is_marked"`
	MarkedIntent *int64 `json:"marked_intent,omitempty"`
	IsIgnored    bool   `json:"is_ignored"`
}

type VisitRecordsQueryResult struct {
	Data        []*VisitRecordsData
	TotalSize   int64
	IgnoredSize int64
	MarkedSize  int64
}

type VisitRecordsExportData struct {
	VisitRecordsCommon
	CustomInfo string `json:"custom_info"`
}

var VisitRecordsTableHeader = map[string][]data.TableHeaderItem{
	"zh-cn": []data.TableHeaderItem{
		data.TableHeaderItem{
			Text: "会话ID",
			ID:   common.VisitRecordsMetricSessionID,
		},
		data.TableHeaderItem{
			Text: "多轮ID",
			ID:   common.VisitRecordsMetricTESessionID,
		},
		data.TableHeaderItem{
			Text: "用户ID",
			ID:   common.VisitRecordsMetricUserID,
		},
		data.TableHeaderItem{
			Text: "用户问题",
			ID:   common.VisitRecordsMetricUserQ,
		},
		data.TableHeaderItem{
			Text: "信心分数",
			ID:   common.VisitRecordsMetricScore,
		},
		data.TableHeaderItem{
			Text: "标准问题",
			ID:   common.VisitRecordsMetricStdQ,
		},
		data.TableHeaderItem{
			Text: "机器人回答",
			ID:   common.VisitRecordsMetricAnswer,
		},
		data.TableHeaderItem{
			Text: "访问时间",
			ID:   common.VisitRecordsMetricLogTime,
		},
		data.TableHeaderItem{
			Text: "情感",
			ID:   common.VisitRecordsMetricEmotion,
		},
		data.TableHeaderItem{
			Text: "情感分数",
			ID:   common.VisitRecordsMetricEmotionScore,
		},
		data.TableHeaderItem{
			Text: "意图",
			ID:   common.VisitRecordsMetricIntent,
		},
		data.TableHeaderItem{
			Text: "意图分数",
			ID:   common.VisitRecordsMetricIntentScore,
		},
		data.TableHeaderItem{
			Text: "出话模块",
			ID:   common.VisitRecordsMetricModule,
		},
		data.TableHeaderItem{
			Text: "出话来源",
			ID:   common.VisitRecordsMetricSource,
		},
		data.TableHeaderItem{
			Text: "标准问题分类",
			ID:   common.VisitRecordsMetricFaqCategoryName,
		},
		data.TableHeaderItem{
			Text: "标准问题标签",
			ID:   common.VisitRecordsMetricFaqRobotTagName,
		},
		data.TableHeaderItem{
			Text: "反馈选择",
			ID:   common.VisitRecordsMetricFeedback,
		},
		data.TableHeaderItem{
			Text: "反馈文字",
			ID:   common.VisitRecordsMetricCustomFeedback,
		},
		data.TableHeaderItem{
			Text: "反馈时间",
			ID:   common.VisitRecordsMetricFeedbackTime,
		},
		data.TableHeaderItem{
			Text: "出话阈值",
			ID:   common.VisitRecordsMetricThreshold,
		},
		data.TableHeaderItem{
			Text: "响应时间",
			ID:   common.VisitRecordsMetricTSpan,
		},
	},
	"zh-tw": []data.TableHeaderItem{
		data.TableHeaderItem{
			Text: "會話ID",
			ID:   common.VisitRecordsMetricSessionID,
		},
		data.TableHeaderItem{
			Text: "多輪場景ID",
			ID:   common.VisitRecordsMetricTESessionID,
		},
		data.TableHeaderItem{
			Text: "用戶ID",
			ID:   common.VisitRecordsMetricUserID,
		},
		data.TableHeaderItem{
			Text: "用戶問題",
			ID:   common.VisitRecordsMetricUserQ,
		},
		data.TableHeaderItem{
			Text: "信心分數",
			ID:   common.VisitRecordsMetricScore,
		},
		data.TableHeaderItem{
			Text: "標準問題",
			ID:   common.VisitRecordsMetricStdQ,
		},
		data.TableHeaderItem{
			Text: "機器人回答",
			ID:   common.VisitRecordsMetricAnswer,
		},
		data.TableHeaderItem{
			Text: "訪問時間",
			ID:   common.VisitRecordsMetricLogTime,
		},
		data.TableHeaderItem{
			Text: "情感",
			ID:   common.VisitRecordsMetricEmotion,
		},
		data.TableHeaderItem{
			Text: "情感分數",
			ID:   common.VisitRecordsMetricEmotionScore,
		},
		data.TableHeaderItem{
			Text: "意圖",
			ID:   common.VisitRecordsMetricIntent,
		},
		data.TableHeaderItem{
			Text: "意圖分數",
			ID:   common.VisitRecordsMetricIntentScore,
		},
		data.TableHeaderItem{
			Text: "出話模組",
			ID:   common.VisitRecordsMetricModule,
		},
		data.TableHeaderItem{
			Text: "出話來源",
			ID:   common.VisitRecordsMetricSource,
		},
		data.TableHeaderItem{
			Text: "標準問題分類",
			ID:   common.VisitRecordsMetricFaqCategoryName,
		},
		data.TableHeaderItem{
			Text: "標準問題標籤",
			ID:   common.VisitRecordsMetricFaqRobotTagName,
		},
		data.TableHeaderItem{
			Text: "反饋選擇",
			ID:   common.VisitRecordsMetricFeedback,
		},
		data.TableHeaderItem{
			Text: "反饋文字",
			ID:   common.VisitRecordsMetricCustomFeedback,
		},
		data.TableHeaderItem{
			Text: "反饋時間",
			ID:   common.VisitRecordsMetricFeedbackTime,
		},
		data.TableHeaderItem{
			Text: "出話閾值",
			ID:   common.VisitRecordsMetricThreshold,
		},
		data.TableHeaderItem{
			Text: "響應時間",
			ID:   common.VisitRecordsMetricTSpan,
		},
	},
}

func GetVisitRecordsExportHeader(locale string) []string {
	var VisitRecordsExportHeader = []string{
		localemsg.Get(locale, "sessionId"),
		localemsg.Get(locale, "taskEngineId"),
		localemsg.Get(locale, "userId"),
		localemsg.Get(locale, "userQ"),
		localemsg.Get(locale, "FAQ"),
		localemsg.Get(locale, "robotAnswer"),
		localemsg.Get(locale, "robotRawAnswer"),
		localemsg.Get(locale, "matchScore"),
		localemsg.Get(locale, "module"),
		localemsg.Get(locale, "source"),
		localemsg.Get(locale, "logTime"),
		localemsg.Get(locale, "emotionCol"),
		localemsg.Get(locale, "emotionScore"),
		localemsg.Get(locale, "intent"),
		localemsg.Get(locale, "intentScore"),
		localemsg.Get(locale, "customInfo"),
		localemsg.Get(locale, "FAQCategory"),
		localemsg.Get(locale, "FAQLabel"),
		localemsg.Get(locale, "feedback"),
		localemsg.Get(locale, "customFeedback"),
		localemsg.Get(locale, "feedbackTime"),
		localemsg.Get(locale, "threshold"),
		localemsg.Get(locale, "respondTime"),
	}
	return VisitRecordsExportHeader
}

type VisitRecordsExportResponse struct {
	ExportID string `json:"export_id"`
}

type VisitRecordsExportStatusResponse struct {
	Status string `json:"status"`
}
