package data

import (
	"strconv"
	"time"
)

const (
	CallStatusComplete       = 1
	CallStatusOngoing        = 0
	CallStatusTranserToHuman = -1
	CallStatusTimeout        = -2
	CallStatusCancel         = -3
)

const (
	CallStatsTypeTime    = "time"
	CallStatsTypeAnswers = "answers"
)

const (
	CallsMetricTotals        = "totals"
	CallsMetricCompletes     = "completes"
	CallsMetricCompletesRate = "completes_rate"
	CallsMetricToHumans      = "to_humans"
	CallsMetricToHumansRate  = "to_humans_rate"
	CallsMetricTimeouts      = "timeouts"
	CallsMetricTimeoutsRate  = "timeouts_rate"
	CallsMetricCancels       = "cancels"
	CallsMetricCancelsRate   = "cancels_rate"
	CallsMetricUnknowns      = "unknowns"
)

type CallStatsQuery struct {
	CommonQuery
	AggInterval string
}

type CallStatsResponse struct {
	TableHeader []TableHeaderItem `json:"table_header"`
	Data        CallStatData      `json:"data"`
	Total       CallStatsTotal    `json:"total"`
}

type CallStatData struct {
	Quantities []CallStatsQuantity `json:"quantities"`
	Type       string              `json:"type"`
	Name       string              `json:"name"`
}

type CallStatsTotal struct {
	CallStatsQ
	TimeText string `json:"time_txt"`
	Time     string `json:"time,omitempty"`
}

var CallStatsTableHeader = []TableHeaderItem{
	TableHeaderItem{
		Text: "统计项",
		ID:   "time_txt",
	},
	TableHeaderItem{
		Text: "总场景数",
		ID:   CallsMetricTotals,
	},
	TableHeaderItem{
		Text: "场景完成数",
		ID:   CallsMetricCompletes,
	},
	TableHeaderItem{
		Text: "场景完成率",
		ID:   CallsMetricCompletesRate,
	},
	TableHeaderItem{
		Text: "场景转人工数",
		ID:   CallsMetricToHumans,
	},
	TableHeaderItem{
		Text: "场景转人工率",
		ID:   CallsMetricToHumansRate,
	},
	TableHeaderItem{
		Text: "场景中断数",
		ID:   CallsMetricTimeouts,
	},
	TableHeaderItem{
		Text: "场景中断率",
		ID:   CallsMetricTimeoutsRate,
	},
	TableHeaderItem{
		Text: "用户拒绝场景数",
		ID:   CallsMetricCancels,
	},
	TableHeaderItem{
		Text: "用户拒绝场景率",
		ID:   CallsMetricCancelsRate,
	},
	TableHeaderItem{
		Text: "未提供型号尺寸数",
		ID:   CallsMetricUnknowns,
	},
}

type CallStatsQuantity struct {
	CallStatsQ
	TimeText string `json:"time_txt"`
	Time     string `json:"time"`
}

type CallStatsQuantities []CallStatsQuantity

func (q CallStatsQuantities) Len() int {
	return len(q)
}

func (q CallStatsQuantities) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q CallStatsQuantities) Less(i, j int) bool {
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

type ToHumanAnswer struct {
	Answer string `json:"answer"`
	Count  int64  `json:"count"`
}

type ToHumanAnswers []ToHumanAnswer

func (answers ToHumanAnswers) Len() int {
	return len(answers)
}

func (answers ToHumanAnswers) Swap(i, j int) {
	answers[i], answers[j] = answers[j], answers[i]
}

func (answers ToHumanAnswers) Less(i, j int) bool {
	return answers[i].Count < answers[j].Count
}

type TopToHumanAnswersResponse struct {
	Answers []TopToHumanAnswer `json:"answers"`
}

type TopToHumanAnswer struct {
	ToHumanAnswer
	Rank int `json:"rank"`
}

type CallStatsQ struct {
	Totals        int64  `json:"totals"`
	Completes     int64  `json:"completes"`
	CompletesRate string `json:"completes_rate"`
	ToHumans      int64  `json:"to_humans"`
	ToHumansRate  string `json:"to_humans_rate"`
	Timeouts      int64  `json:"timeouts"`
	TimeoutsRate  string `json:"timeouts_rate"`
	Cancels       int64  `json:"cancels"`
	CancelsRate   string `json:"cancels_rate"`
	Unknowns      int64  `json:"unknowns"`
}

func NewCallStatsQ() *CallStatsQ {
	return &CallStatsQ{
		Totals:        0,
		Completes:     0,
		CompletesRate: "N/A",
		ToHumans:      0,
		ToHumansRate:  "N/A",
		Timeouts:      0,
		TimeoutsRate:  "N/A",
		Cancels:       0,
		CancelsRate:   "N/A",
		Unknowns:      0,
	}
}

type CallStatsQueryHandler func(query CallStatsQuery) (map[string]interface{}, error)
