package v1

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
)

type TEVisitRecordsRequest struct {
	StartTime    int64    `json:"start_time"`
	EndTime      int64    `json:"end_time"`
	ScenarioName *string  `json:"scenario_name,omitempty"`
	Platforms    []string `json:"platforms,omitempty"`
	Genders      []string `json:"genders,omitempty"`
	UserID       *string  `json:"uid,omitempty"`
	Feedback     *string  `json:"feedback,omitempty"`
	Page         *int64   `json:"page,omitempty"`
	Limit        *int64   `json:"limit,omitempty"`
}

type TEVisitRecordsQuery struct {
	data.CommonQuery
	ScenarioName *string
	Platforms    []string
	Genders      []string
	UserID       *string
	Feedback     *string
	From         int64
	Limit        int64
}

type TEVisitRecordsResponse struct {
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        []*TEVisitRecordsData  `json:"data"`
	Limit       int64                  `json:"limit"`
	Page        int64                  `json:"page"`
	TotalSize   int64                  `json:"total_size"`
}

type TEVisitRecordsDataBase struct {
	TESessionID    string `json:"taskengine_session_id"`
	SessionID      string `json:"session_id"`
	UserID         string `json:"user_id"`
	ScenarioID     string `json:"scenario_id"`
	ScenarioName   string `json:"scenario_name"`
	LastNodeID     string `json:"last_node_id"`
	LastNodeName   string `json:"last_node_name"`
	Feedback       string `json:"feedback"`
	CustomFeedback string `json:"custom_feedback"`
}

type TEVisitRecordsRawData struct {
	TEVisitRecordsDataBase
	TriggerTime  int64 `json:"trigger_time"`
	FinishTime   int64 `json:"finish_time"`
	FeedbackTime int64 `json:"feedback_time"`
}

type TEVisitRecordsData struct {
	TEVisitRecordsDataBase
	TriggerTime  string `json:"trigger_time"`
	FinishTime   string `json:"finish_time"`
	FeedbackTime string `json:"feedback_time"`
}

var TEVisitRecordsTableHeader = map[string][]data.TableHeaderItem{
	"zh-cn": []data.TableHeaderItem{
		data.TableHeaderItem{
			Text: "场景会话ID",
			ID:   common.TEVisitRecordsMetricTESessionID,
		},
		data.TableHeaderItem{
			Text: "会话ID",
			ID:   common.TEVisitRecordsMetricSessionID,
		},
		data.TableHeaderItem{
			Text: "用户ID",
			ID:   common.TEVisitRecordsMetricUserID,
		},
		data.TableHeaderItem{
			Text: "场景 ID",
			ID:   common.TEVisitRecordsMetricScenarioID,
		},
		data.TableHeaderItem{
			Text: "场景名称",
			ID:   common.TEVisitRecordsMetricScenarioName,
		},
		data.TableHeaderItem{
			Text: "最终结点ID",
			ID:   common.TEVisitRecordsMetricLastNodeID,
		},
		data.TableHeaderItem{
			Text: "最终结点名称",
			ID:   common.TEVisitRecordsMetricLastNodeName,
		},
		data.TableHeaderItem{
			Text: "触发时间",
			ID:   common.TEVisitRecordsMetricTriggerTime,
		},
		data.TableHeaderItem{
			Text: "完成时间",
			ID:   common.TEVisitRecordsMetricFinishTime,
		},
		data.TableHeaderItem{
			Text: "反馈选择",
			ID:   common.TEVisitRecordsMetricFeedback,
		},
		data.TableHeaderItem{
			Text: "反馈文字",
			ID:   common.TEVisitRecordsMetricCustomFeedback,
		},
		data.TableHeaderItem{
			Text: "反馈时间",
			ID:   common.TEVisitRecordsMetricFeedbackTime,
		},
	},
	"zh-tw": []data.TableHeaderItem{
		data.TableHeaderItem{
			Text: "場景會話ID",
			ID:   common.TEVisitRecordsMetricTESessionID,
		},
		data.TableHeaderItem{
			Text: "會話ID",
			ID:   common.TEVisitRecordsMetricSessionID,
		},
		data.TableHeaderItem{
			Text: "用戶ID",
			ID:   common.TEVisitRecordsMetricUserID,
		},
		data.TableHeaderItem{
			Text: "場景 ID",
			ID:   common.TEVisitRecordsMetricScenarioID,
		},
		data.TableHeaderItem{
			Text: "場景名稱",
			ID:   common.TEVisitRecordsMetricScenarioName,
		},
		data.TableHeaderItem{
			Text: "最終結點ID",
			ID:   common.TEVisitRecordsMetricLastNodeID,
		},
		data.TableHeaderItem{
			Text: "最終結點名稱",
			ID:   common.TEVisitRecordsMetricLastNodeName,
		},
		data.TableHeaderItem{
			Text: "觸發時間",
			ID:   common.TEVisitRecordsMetricTriggerTime,
		},
		data.TableHeaderItem{
			Text: "完成時間",
			ID:   common.TEVisitRecordsMetricFinishTime,
		},
		data.TableHeaderItem{
			Text: "反饋選擇",
			ID:   common.TEVisitRecordsMetricFeedback,
		},
		data.TableHeaderItem{
			Text: "反饋文字",
			ID:   common.TEVisitRecordsMetricCustomFeedback,
		},
		data.TableHeaderItem{
			Text: "反饋時間",
			ID:   common.TEVisitRecordsMetricFeedbackTime,
		},
	},
}

var TEVisitRecordsExportHeader = []string{
	"场景会话ID",
	"会话ID",
	"用户ID",
	"场景ID",
	"场景名称",
	"最终结点ID",
	"最终结点名称",
	"触发时间",
	"完成时间",
	"反馈选择",
	"反馈文字",
	"反馈时间",
}

type TEVisitRecordsExportResponse struct {
	ExportID string `json:"export_id"`
}

type TEVisitRecordsExportStatusResponse struct {
	Status string `json:"status"`
}
