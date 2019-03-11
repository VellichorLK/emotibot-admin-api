package v1

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
)

type FeedbacksQuery struct {
	data.CommonQuery
	Type     string
	Platform string
	Gender   string
}

type FeedbacksResponse struct {
	TableHeader []data.TableHeaderItem `json:"table_header"`
	Data        FeedbacksData          `json:"data"`
}

type FeedbacksData struct {
	AvgRating float64          `json:"avg_rating"`
	Ratings   map[string]int64 `json:"ratings"`
	Feedbacks map[string]int64 `json:"feedbacks"`
}

type FeedbackRating struct {
	Rating string
	Count  int64
}

type FeedbackRatings []FeedbackRating

func (f FeedbackRatings) Len() int {
	return len(f)
}

func (f FeedbackRatings) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FeedbackRatings) Less(i, j int) bool {
	return f[i].Count < f[j].Count
}

type FeedbackCount struct {
	Feedback string
	Count    int64
}

type FeedbackCounts []FeedbackCount

func (f FeedbackCounts) Len() int {
	return len(f)
}

func (f FeedbackCounts) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FeedbackCounts) Less(i, j int) bool {
	return f[i].Count < f[j].Count
}

var FeedbacksTableHeader = []data.TableHeaderItem{
	data.TableHeaderItem{
		Text: "统计项",
		ID:   "time_txt",
	},
	data.TableHeaderItem{
		Text: "平均满意度",
		ID:   common.FeedbacksMetricAvgRating,
	},
	data.TableHeaderItem{
		Text: "满意度",
		ID:   common.FeedbacksMetricRatings,
	},
	data.TableHeaderItem{
		Text: "回馈数",
		ID:   common.FeedbacksMetricFeedbacks,
	},
}

type RatingAverageInfo struct {
	Average float64         `json:"average"`
	Count   map[int64]int64 `json:"count"`
}
