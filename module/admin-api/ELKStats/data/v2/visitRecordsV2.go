package v2

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
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
	SessionID   string  `json:"session_id"`
	TESessionID string  `json:"taskengine_session_id"`
	UserID      string  `json:"user_id"`
	UserQ       string  `json:"user_q"`
	Score       float64 `json:"score"`
	StdQ        string  `json:"std_q"`
	LogTime     string  `json:"log_time"`
	Emotion     string  `json:"emotion"`
	Intent      string  `json:"intent"`
	Module      string  `json:"module"`
}

type VisitRecordsRawData struct {
	VisitRecordsDataBase
	EmotionScore   float64                `json:"emotion_score"`
	IntentScore    float64                `json:"intent_score"`
	UniqueID       string                 `json:"unique_id"`
	Answer         []data.Answer          `json:"answer"`
	CustomInfo     map[string]interface{} `json:"custom_info"`
	Source         string                 `json:"source"`
	Note           string                 `json:"note"`
	IsMarked       bool                   `json:"is_marked"`
	IsIgnored      bool                   `json:"is_ignored"`
	FaqCategoryID  int64                  `json:"faq_cat_id"`
	FaqRobotTagIDs []int64                `json:"faq_robot_tag_id"`
	Feedback       string                 `json:"feedback"`
	CustomFeedback string                 `json:"custom_feedback"`
	FeedbackTime   int64                  `json:"feedback_time"`
	Threshold      int64                  `json:"threshold"`
}

type VisitRecordsCommon struct {
	VisitRecordsDataBase
	Answer           string  `json:"answer"`
	FaqCategoryName  string  `json:"faq_cat_name"`
	FaqRobotTagNames string  `json:"faq_robot_tag_name"`
	FaqCategoryID    int64   `json:"-"` // Redundant field for FAQ category name conversion
	FaqRobotTagIDs   []int64 `json:"-"` // Redundant field for FAQ robot tag name conversion
	Feedback         string  `json:"feedback"`
	CustomFeedback   string  `json:"custom_feedback"`
	FeedbackTime     string  `json:"feedback_time"`
	Threshold        int64   `json:"threshold"`
}
type VisitRecordsData struct {
	VisitRecordsCommon
	UniqueID  string `json:"id"`
	IsMarked  bool   `json:"is_marked"`
	IsIgnored bool   `json:"is_ignored"`
}

type VisitRecordsQueryResult struct {
	Data        []*VisitRecordsData
	TotalSize   int64
	IgnoredSize int64
	MarkedSize  int64
}

type VisitRecordsExportData struct {
	VisitRecordsCommon
	Source       string  `json:"source"`
	EmotionScore float64 `json:"emotion_score"`
	IntentScore  float64 `json:"intent_score"`
	CustomInfo   string  `json:"custom_info"`
}

var VisitRecordsTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "会话ID",
		ID:   common.VisitRecordsMetricSessionID,
	},
	data.TableHeaderItem{
		Text: "多輪場景ID",
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
		Text: "匹配分数",
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
		Text: "意图",
		ID:   common.VisitRecordsMetricIntent,
	},
	data.TableHeaderItem{
		Text: "出话模组",
		ID:   common.VisitRecordsMetricModule,
	},
	data.TableHeaderItem{
		Text: "FAQ 分类",
		ID:   common.VisitRecordsMetricFaqCategoryName,
	},
	data.TableHeaderItem{
		Text: "FAQ 标签",
		ID:   common.VisitRecordsMetricFaqRobotTagName,
	},
	data.TableHeaderItem{
		Text: "反馈",
		ID:   common.VisitRecordsMetricFeedback,
	},
	data.TableHeaderItem{
		Text: "客制化反馈",
		ID:   common.VisitRecordsMetricCustomFeedback,
	},
	data.TableHeaderItem{
		Text: "反馈时间",
		ID:   common.VisitRecordsMetricFeedbackTime,
	},
	data.TableHeaderItem{
		Text: "出话阀值",
		ID:   common.VisitRecordsMetricThreshold,
	},
}

var VisitRecordsExportHeader = []string{
	"会话ID",
	"多輪場景ID",
	"用户ID",
	"用户问题",
	"标准问题",
	"机器人回答",
	"匹配分数",
	"出话模组",
	"出话来源",
	"访问时间",
	"情感",
	"情感分数",
	"意图",
	"意图分数",
	"客制化资讯",
	"FAQ 分类",
	"FAQ 标签",
	"反馈",
	"客制化反馈",
	"反馈时间",
	"出话阀值",
}

type VisitRecordsExportResponse struct {
	ExportID string `json:"export_id"`
}

type VisitRecordsExportStatusResponse struct {
	Status string `json:"status"`
}
