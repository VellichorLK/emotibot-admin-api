package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func genAvgEmotion(date int64, totalCount uint64, sumDuration uint64,
	ch1Count uint64, ch1Score float64, ch2Count uint64, ch2Score float64) *AvgEmotion {

	ae := &AvgEmotion{}
	ae.Total = totalCount

	t := time.Unix(date, 0)
	ae.Date = fmt.Sprintf("%d/%02d/%02d", t.Year(), t.Month(), t.Day())

	precesion := 1

	if totalCount > 0 {
		ae.AvgDuration = uint64(float64(sumDuration) / float64(totalCount))
	}
	if ch1Count > 0 {
		ae.AvgCh1Anger = GetFloatPrecesion(float64(ch1Score)/float64(ch1Count), precesion)
	}
	if ch2Count > 0 {
		ae.AvgCh2Anger = GetFloatPrecesion(float64(ch2Score)/float64(ch2Count), precesion)
	}

	return ae
}

func genSimpleAvgEmotion(date int64, sc *ScoreCount) *SimpleAvgEmotion {
	sae := &SimpleAvgEmotion{}
	t := time.Unix(date, 0)
	sae.Date = fmt.Sprintf("%d/%02d/%02d", t.Year(), t.Month(), t.Day())

	precesion := 1
	if sc.count1 > 0 {
		sae.AvgCh1Anger = GetFloatPrecesion(sc.score1/float64(sc.count1), precesion)
	}
	if sc.count2 > 0 {
		sae.AvgCh2Anger = GetFloatPrecesion(sc.score2/float64(sc.count2), precesion)
	}

	return sae
}

func addGroupAvgEmotion(date int64, group map[string]*ScoreCount, groups map[string][]*SimpleAvgEmotion) {
	for tag, scores := range group {
		var saes []*SimpleAvgEmotion
		var ok bool
		if saes, ok = groups[tag]; !ok {
			saes = make([]*SimpleAvgEmotion, 0)
		}

		sae := genSimpleAvgEmotion(date, scores)
		saes = append(saes, sae)

		fmt.Printf("Date:%s, count1:%d, avg1:%v, count2:%d, avg2:%v, tag:%s\n",
			sae.Date, scores.count1, sae.AvgCh1Anger, scores.count2, sae.AvgCh2Anger, tag)

		groups[tag] = saes
	}

}

func genDefaultWhereStates(_t1 uint64, _t2 uint64, appid string) ([2][]WhereStates, []interface{}) {
	whereStates := [2][]WhereStates{
		{
			{name: NAPPID, compare: "="},
			{name: NANARES, compare: "="},
			{name: NFILET, compare: ">="},
			{name: NFILET, compare: "<="},
		},
		{
			{name: NEMOTYPE, compare: "="},
		},
	}

	params := make([]interface{}, 0, 5)
	params = append(params, appid)
	params = append(params, 1) //analysis_result = 1
	params = append(params, _t1)
	e := AddTimeUnit(int64(_t2), Day) - 1
	params = append(params, e)
	params = append(params, 1) // emotion_type = 1 means anger

	return whereStates, params
}

func genDefaultOrderStates() [2][]string {
	orderStates := [2][]string{{NFILET, NID}}
	return orderStates
}

func genGroupReport(saeMap map[string][]*SimpleAvgEmotion) *GroupReport {

	if saeMap == nil {
		return nil
	}

	numOfGrp := len(saeMap)

	gr := &GroupReport{}
	gr.Group = make([]*GroupAvgEmotion, 0, numOfGrp)
	for filter, sae := range saeMap {
		gve := &GroupAvgEmotion{}
		gve.Tag = filter
		gve.Data = sae
		gr.Group = append(gr.Group, gve)
	}

	return gr

}

func groupAvgEmotion(_t1, _t2 uint64, appid string, filter string) (map[string][]*SimpleAvgEmotion, error) {
	selectColumns := [2][]string{{NID, NFILET, filter}, {NCHANNEL, NSCORE}}
	whereStates, params := genDefaultWhereStates(_t1, _t2, appid)
	orderStates := genDefaultOrderStates()

	rows, err := getEmotionData(selectColumns, whereStates, orderStates, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	t1 := int64(_t1)
	timeUnit := Week
	timeUnitInSec := int64(7 * 24 * 60 * 60)
	nextDay := AddTimeUnit(t1, timeUnit)

	var createdTime int64
	var lastID uint64
	var sc *ScoreCount
	var ok bool

	group := make(map[string]*ScoreCount)
	groups := make(map[string][]*SimpleAvgEmotion)

	totalCount := 0

	for rows.Next() {
		var id uint64
		var ch int8
		var score float64
		var tag sql.NullString

		err := rows.Scan(&id, &createdTime, &tag, &ch, &score)

		if err != nil {
			return nil, err
		}

		if createdTime >= nextDay {
			fmt.Printf("created_time:%v, nextDay:%v, id:%v\n", createdTime, nextDay, id)
			fmt.Printf("total count:%d\n", totalCount)
			//fmt.Printf("len:%v\n", len(group))
			addGroupAvgEmotion(nextDay-timeUnitInSec, group, groups)
			group = make(map[string]*ScoreCount)
			nextDay = AddTimeUnit(nextDay, timeUnit)

			fmt.Println("-------------------------------")
		}

		if lastID != id {
			lastID = id
			totalCount++
			if sc, ok = group[tag.String]; !ok {
				sc = &ScoreCount{}
				group[tag.String] = sc
			}
		}

		switch ch {
		case 1:
			sc.count1++
			sc.score1 += score
		case 2:
			sc.count2++
			sc.score2 += score
		default:
			return nil, errors.New("channel number out of bound")
		}
	}

	addGroupAvgEmotion(nextDay-timeUnitInSec, group, groups)
	return groups, nil
}

func dailyAvgEmotion(_t1 uint64, _t2 uint64, appid string) ([]*AvgEmotion, error) {

	selectColumns := [2][]string{{NID, NRDURATION, NFILET}, {NCHANNEL, NSCORE}}
	whereStates, params := genDefaultWhereStates(_t1, _t2, appid)
	orderStates := genDefaultOrderStates()

	rows, err := getEmotionData(selectColumns, whereStates, orderStates, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	t1 := int64(_t1)
	nextDay := AddTimeUnit(t1, Day)
	var totalCount uint64      //total audio file in one day
	var score1, score2 float64 //anger score of channel1 and channel2
	var count1, count2 uint64  //count of channel1 and channel2
	var sumDuration uint64
	var createdTime int64
	var lastID uint64

	avgEmotions := make([]*AvgEmotion, 0)

	for rows.Next() {
		var id, duration uint64
		var ch int8
		var score float64

		err := rows.Scan(&id, &duration, &createdTime, &ch, &score)

		if err != nil {
			return nil, err
		}

		if createdTime >= nextDay {
			ae := genAvgEmotion(nextDay-1, totalCount, sumDuration, count1, score1, count2, score2)

			avgEmotions = append(avgEmotions, ae)

			totalCount, sumDuration = 0, 0
			count1, count2 = 0, 0
			score1, score2 = 0, 0

			nextDay = AddTimeUnit(nextDay, Day)
		}

		if lastID != id {
			totalCount++
			sumDuration += (duration / 1000)
			lastID = id
		}

		switch ch {
		case 1:
			count1++
			score1 += score
		case 2:
			count2++
			score2 += score
		default:
			return nil, errors.New("channel number out of bound")
		}
	}

	ae := genAvgEmotion(nextDay-1, totalCount, sumDuration, count1, score1, count2, score2)
	avgEmotions = append(avgEmotions, ae)

	return avgEmotions, nil
}
