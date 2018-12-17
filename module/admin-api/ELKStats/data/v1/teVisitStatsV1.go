package v1

import (
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	common "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
)

type TEVisitStatsQuery struct {
	data.StatsQuery
	ScenarioID   string
	ScenarioName string
}

type TEVisitStatsResponse struct {
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        TEVisitStatsData       `json:"data"`
	Total       TEVisitStatsTotal      `json:"total"`
}

type TEVisitStatsData struct {
	TEVisitStatsQuantities []TEVisitStatsQuantity `json:"quantities"`
	Type                   string                 `json:"type"`
}

type TEVisitStatsTotal struct {
	TEVisitStatsQ
}

type TEVisitStatsTagResponse struct {
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        []TEVisitStatsTagData  `json:"data"`
	Total       TEVisitStatsQ          `json:"total"`
}

type TEVisitStatsTagData struct {
	Q    TEVisitStatsQ `json:"q"`
	ID   string        `json:"id"`
	Name string        `json:"name"`
}

type TEVisitStatsQuantity struct {
	TEVisitStatsQ
	TimeText string `json:"time_txt"`
	Time     string `json:"time"`
}

type TEVisitStatsQuantities []TEVisitStatsQuantity

func (q TEVisitStatsQuantities) Len() int {
	return len(q)
}

func (q TEVisitStatsQuantities) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q TEVisitStatsQuantities) Less(i, j int) bool {
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

var TEVisitStatsTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "统计项",
		ID:   "time_txt",
	},
	data.TableHeaderItem{
		Text: "触发数",
		ID:   common.TEVisitStatsMetricTriggers,
	},
	data.TableHeaderItem{
		Text: "未完成数",
		ID:   common.TEVisitStatsMetricUnfinished,
	},
}

var TEVisitStatsTagTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "触发数",
		ID:   common.TEVisitStatsMetricTriggers,
	},
	data.TableHeaderItem{
		Text: "未完成数",
		ID:   common.TEVisitStatsMetricUnfinished,
	},
}

type TEVisitStatsQ struct {
	Triggers   int64 `json:"triggers"`
	Unfinished int64 `json:"unfinished"`
}

type TEVisitStatsQueryHandler func(query TEVisitStatsQuery) (map[string]interface{}, error)
