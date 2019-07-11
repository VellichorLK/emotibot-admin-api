package common

const (
	RecordsDefaultPage      = 1
	RecordsDefaultPageLimit = 20
)

const (
	CategoryBusiness = "business"
	CategoryChat     = "chat"
	CategoryOther    = "other"
)

const (
	VisitRecordsMetricSessionID       = "session_id"
	VisitRecordsMetricUserID          = "user_id"
	VisitRecordsMetricUserQ           = "user_q"
	VisitRecordsMetricStdQ            = "std_q"
	VisitRecordsMetricAnswer          = "answer"
	VisitRecordsMetricRawAnswer       = "raw_answer"
	VisitRecordsMetricModule          = "module"
	VisitRecordsMetricEmotion         = "emotion"
	VisitRecordsMetricEmotionScore    = "emotion_score"
	VisitRecordsMetricIntent          = "intent"
	VisitRecordsMetricIntentScore     = "intent_score"
	VisitRecordsMetricLogTime         = "log_time"
	VisitRecordsMetricScore           = "score"
	VisitRecordsMetricCustomInfo      = "custom_info"
	VisitRecordsMetricSource          = "source"
	VisitRecordsMetricNote            = "note"
	VisitRecordsMetricQType           = "qtype"
	VisitRecordsMetricTESessionID     = "taskengine_session_id"
	VisitRecordsMetricFaqCategoryName = "faq_cat_name"
	VisitRecordsMetricFaqRobotTagName = "faq_robot_tag_name"
	VisitRecordsMetricFeedback        = "feedback"
	VisitRecordsMetricCustomFeedback  = "custom_feedback"
	VisitRecordsMetricFeedbackTime    = "feedback_time"
	VisitRecordsMetricThreshold       = "threshold"
	VisitRecordsMetricTSpan           = "tspan"

	// ccs
	VisitRecordsMetricUnique      = "unique_id"
	VisitRecordsMetricRuleIDs     = "rule_ids"
	VisitRecordsMetricTaskID      = "task_id"
	VisitRecordsMetricRawResponse = "raw_response"
	VisitTagsMetricAnswers		  = "answers"
	VisitTagsMetricAIModule		  = "ai_module"
	VisitTagsMetricCaseId		  = "case_id"
	VisitTagsMetricAppId		  = "app_id"
	VisitTagsMetricDataset		  = "dataset"
)

type VisitRecordsExportResponse struct {
	ExportID string `json:"export_id"`
}

type VisitRecordsExportStatusResponse struct {
	Status string `json:"status"`
}
