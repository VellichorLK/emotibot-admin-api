package v1

import (
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
)

type CallStatsQuery struct {
	data.CommonQuery
	AggInterval string
}

type CallStatsResponse struct {
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        CallStatData           `json:"data"`
	Total       CallStatsTotal         `json:"total"`
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

var CallStatsTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "统计项",
		ID:   "time_txt",
	},
	data.TableHeaderItem{
		Text: "总场景数",
		ID:   common.CallsMetricTotals,
	},
	data.TableHeaderItem{
		Text: "场景完成数",
		ID:   common.CallsMetricCompletes,
	},
	data.TableHeaderItem{
		Text: "场景完成率",
		ID:   common.CallsMetricCompletesRate,
	},
	data.TableHeaderItem{
		Text: "场景转人工数",
		ID:   common.CallsMetricToHumans,
	},
	data.TableHeaderItem{
		Text: "场景转人工率",
		ID:   common.CallsMetricToHumansRate,
	},
	data.TableHeaderItem{
		Text: "场景中断数",
		ID:   common.CallsMetricTimeouts,
	},
	data.TableHeaderItem{
		Text: "场景中断率",
		ID:   common.CallsMetricTimeoutsRate,
	},
	data.TableHeaderItem{
		Text: "用户拒绝场景数",
		ID:   common.CallsMetricCancels,
	},
	data.TableHeaderItem{
		Text: "用户拒绝场景率",
		ID:   common.CallsMetricCancelsRate,
	},
	data.TableHeaderItem{
		Text: "未提供型号尺寸数",
		ID:   common.CallsMetricUnknowns,
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
