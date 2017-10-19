package main

import (
	"errors"
	"fmt"
	"handlers"
	"reflect"
	"strconv"
	"time"
)

type Summary struct {
	TotalNum      string
	TotalDuration string
	Channel1Avg   string
	Channel2Avg   string
}
type Detail struct {
	TimeRange string
	Summ      Summary
}

type Stat struct {
	Channel1Text string
	Channel2Text string
	Summaries    []*Summary
	Analysis     []*Summary
	Details      []*Detail
}

type Tmpl1 struct {
	TargetDate string
	Stat
}
type Tmpl2 struct {
	TargetDateFrom string
	TargetDateTo   string
	Stat
}

//GetTmplMsg fill the template with data
func FillTmplStruct(lastS *handlers.Statistics, nowS *handlers.Statistics, tmpl interface{}, timeFormat string) error {

	s := reflect.ValueOf(tmpl)
	if s == (reflect.Value{}) {
		return errors.New("error. template is nil")
	}

	s = s.Elem()
	if s == (reflect.Value{}) {
		return errors.New("error. template is not ptr")
	}

	s = s.FieldByName("Stat")
	if s == (reflect.Value{}) {
		return errors.New("error. no Stat struct in template")
	}

	ch1Text := s.FieldByName("Channel1Text")
	if ch1Text == (reflect.Value{}) {
		return errors.New("error. No Channel1Text field in Stat struct")
	}
	ch1Text.SetString("客服")

	ch2Text := s.FieldByName("Channel2Text")
	if ch2Text == (reflect.Value{}) {
		return errors.New("error. No Channel2Text field in Stat struct")
	}
	ch2Text.SetString("客户")

	thisSum := &Summary{TotalNum: strconv.FormatUint(nowS.Count, 10),
		TotalDuration: durationFormat(nowS.Duration),
		Channel1Avg:   strconv.FormatFloat(nowS.Ch1AvgAnger, 'f', 2, 64),
		Channel2Avg:   strconv.FormatFloat(nowS.Ch2AvgAnger, 'f', 2, 64),
	}

	lastSum := &Summary{TotalNum: strconv.FormatUint(lastS.Count, 10),
		TotalDuration: durationFormat(lastS.Duration),
		Channel1Avg:   strconv.FormatFloat(lastS.Ch1AvgAnger, 'f', 2, 64),
		Channel2Avg:   strconv.FormatFloat(lastS.Ch2AvgAnger, 'f', 2, 64),
	}

	diffSum := diffPercent(lastS, nowS)

	sum := s.FieldByName("Summaries")
	if sum == (reflect.Value{}) {
		return errors.New("error. No Summaries field in Stat struct")
	}
	sum.Set(reflect.Append(sum, reflect.ValueOf(thisSum)))

	ana := s.FieldByName("Analysis")
	if ana == (reflect.Value{}) {
		return errors.New("error. No Analysis field in Stat struct")
	}
	ana.Set(reflect.Append(ana, reflect.ValueOf(lastSum)))
	ana.Set(reflect.Append(ana, reflect.ValueOf(thisSum)))
	ana.Set(reflect.Append(ana, reflect.ValueOf(diffSum)))

	det := s.FieldByName("Details")
	if det == (reflect.Value{}) {
		return errors.New("error. No Details field in Stat struct")
	}

	for _, v := range nowS.Data {
		d := statToDetail(v, timeFormat)
		det.Set(reflect.Append(det, reflect.ValueOf(d)))
	}

	return nil
}

func diffPercent(lastS *handlers.Statistics, nowS *handlers.Statistics) *Summary {

	format := "%+.2f%%"
	infinity := "+Inf%"

	var countDiff, durDiff, ch1Diff, ch2Diff string

	if lastS.Count == 0 {
		countDiff = infinity
	} else {
		countDiff = fmt.Sprintf(format, float64((nowS.Count-lastS.Count)*100)/float64(lastS.Count))
	}

	if lastS.Duration == 0 {
		durDiff = infinity
	} else {
		durDiff = fmt.Sprintf(format, float64((nowS.Duration-lastS.Duration)*100)/float64(lastS.Duration))
	}

	if lastS.Ch1AvgAnger == 0 {
		ch1Diff = infinity
	} else {
		ch1Diff = fmt.Sprintf(format, float64((nowS.Ch1AvgAnger-lastS.Ch1AvgAnger)*100)/float64(lastS.Ch1AvgAnger))
	}

	if lastS.Ch2AvgAnger == 0 {
		ch2Diff = infinity
	} else {
		ch2Diff = fmt.Sprintf(format, float64((nowS.Ch2AvgAnger-lastS.Ch2AvgAnger)*100)/float64(lastS.Ch2AvgAnger))
	}

	return &Summary{countDiff, durDiff, ch1Diff, ch2Diff}
}

func statToDetail(s *handlers.StatUnit, timeFormat string) *Detail {
	t := time.Unix(s.From, 0)

	d := &Detail{TimeRange: t.Format(timeFormat),
		Summ: Summary{
			TotalNum:      strconv.FormatUint(s.Count, 10),
			TotalDuration: durationFormat(s.Duration),
			Channel1Avg:   strconv.FormatFloat(s.Ch1AvgAnger, 'f', 2, 64),
			Channel2Avg:   strconv.FormatFloat(s.Ch2AvgAnger, 'f', 2, 64),
		}}

	return d
}

func durationFormat(d uint64) string {
	var h, m, s uint64
	s = d % 60
	m = d / 60
	h = m / 60
	m = m % 60
	return fmt.Sprintf("%02vh:%02vm:%02vs", h, m, s)
}
