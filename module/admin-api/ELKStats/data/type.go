package data

import (
	"time"

	"github.com/olivere/elastic"
)

type CommonQuery struct {
	EnterpriseID string
	AppID        string
	StartTime    time.Time
	EndTime      time.Time
}

type StatsQuery struct {
	CommonQuery
	AggBy       string
	AggInterval string
	AggTagType  string
}

type Tag struct {
	Code string
	Name string
}

type TableHeaderItem struct {
	Text string `json:"text"`
	ID   string `json:"id"`
}

type QueryTag struct {
	Type string
	Text string
}

type QueryTags struct {
	Type  string
	Texts []string
}

type FaqCategoryPath struct {
	ID   int64
	Path string
}

type FaqRobotTag struct {
	ID  int64
	Tag string
}

type Answer struct {
	Type       string        `json:"type"`
	SubType    string        `json:"subType"`
	Value      string        `json:"value"`
	Data       []interface{} `json:"data"`
	ExtendData string        `json:"extendData"`
}

type ExtractExportHitResultHandler func(hit *elastic.SearchHit) (recordPtr interface{}, err error)
type XlsxCreateHandler func(recordPtrs []interface{}, fileName string, locale string) (filePath string, err error)

type ExportTaskOption struct {
	TaskID               string
	Index                string
	BoolQuery            *elastic.BoolQuery
	Source               *elastic.FetchSourceContext
	SortField            string
	ExtractResultHandler ExtractExportHitResultHandler
	XlsxCreateHandler    XlsxCreateHandler
}
