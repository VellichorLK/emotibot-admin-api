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

func genSimpleAvgDuration(date int64, sc *ScoreCount) *SimpleAvgDuration {
	sae := &SimpleAvgDuration{}
	t := time.Unix(date, 0)
	sae.Date = fmt.Sprintf("%d/%02d/%02d", t.Year(), t.Month(), t.Day())

	precesion := 1
	if sc.count1 > 0 {
		avg := GetFloatPrecesion(sc.score1/float64(sc.count1), precesion)
		sae.AvgDuration = int(avg)
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
		groups[tag] = saes
	}
}

func addGroupAvgDuration(date int64, group map[string]*ScoreCount, groups map[string][]*SimpleAvgDuration) {
	for tag, scores := range group {
		var saes []*SimpleAvgDuration
		var ok bool
		if saes, ok = groups[tag]; !ok {
			saes = make([]*SimpleAvgDuration, 0)
		}

		sae := genSimpleAvgDuration(date, scores)
		saes = append(saes, sae)
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
	gr.Group = make([]*GroupAvgData, 0, numOfGrp)
	for filter, sae := range saeMap {
		gve := &GroupAvgData{}
		gve.Tag = filter
		gve.Data = sae
		gr.Group = append(gr.Group, gve)

		var avg float64
		var sum1 float64
		var sum2 float64

		num := len(sae)
		if num > 0 {
			value1 := make([]float64, 0)
			value2 := make([]float64, 0)
			for i := 0; i < num; i++ {
				if sae[i].AvgCh1Anger > 0 {
					value1 = append(value1, sae[i].AvgCh1Anger)
					sum1 += sae[i].AvgCh1Anger
				}
				if sae[i].AvgCh2Anger > 0 {
					value2 = append(value2, sae[i].AvgCh2Anger)
					sum2 += sae[i].AvgCh2Anger
				}
			}
			if len(value1) > 0 {
				avg = sum1 / float64(len(value1))
				gve.Ch1AngerR = GetFloatPrecesion(StdDev(value1)/avg, 2)
				gve.AvgCh1Anger = GetFloatPrecesion(avg, 1)
			}
			if len(value2) > 0 {
				avg = sum2 / float64(len(value2))
				gve.Ch2AngerR = GetFloatPrecesion(StdDev(value2)/avg, 2)
				gve.AvgCh2Anger = GetFloatPrecesion(avg, 1)

			}
		}
	}
	return gr
}

func genGroupDurationReport(saeMap map[string][]*SimpleAvgDuration) *GroupDurationReport {
	if saeMap == nil {
		return nil
	}
	numOfGrp := len(saeMap)

	gr := &GroupDurationReport{}
	gr.Group = make([]*GroupAvgDuration, 0, numOfGrp)
	for filter, sae := range saeMap {
		gve := &GroupAvgDuration{}
		gve.Tag = filter
		gve.Data = sae
		gr.Group = append(gr.Group, gve)

		var avg float64
		var sum uint64

		num := len(sae)

		if num > 0 {
			value := make([]float64, num)
			for i := 0; i < num; i++ {
				sum += uint64(sae[i].AvgDuration)
				value[i] = float64(sae[i].AvgDuration)
			}
			if num > 0 {
				avg = float64(sum) / float64(num)
				gve.AvgDurR = GetFloatPrecesion(StdDev(value)/avg, 2)
				gve.AvgDuration = int(avg)
			}
		}

	}
	return gr
}

func groupAvgEmotion(_t1, _t2 uint64, appid string, filter string) (map[string][]*SimpleAvgEmotion, error) {
	selectColumns := [2][]string{{NID, NFILET, filter}, {NCHANNEL, NSCORE}}
	whereStates, params := genDefaultWhereStates(_t1, _t2, appid)
	orderStates := genDefaultOrderStates()

	rows, err := getEmotionData(selectColumns, whereStates, nil, orderStates, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	t1 := int64(_t1)
	timeUnit := Day
	timeUnitInSec := int64(24 * 60 * 60)
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
			//fmt.Printf("created_time:%v, nextDay:%v, id:%v\n", createdTime, nextDay, id)
			//fmt.Printf("total count:%d\n", totalCount)
			//fmt.Printf("len:%v\n", len(group))
			addGroupAvgEmotion(nextDay-timeUnitInSec, group, groups)
			group = make(map[string]*ScoreCount)
			for createdTime >= nextDay {
				nextDay = AddTimeUnit(nextDay, timeUnit)
			}
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
			sc.scores1 = append(sc.scores1, score)
		case 2:
			sc.count2++
			sc.score2 += score
			sc.scores2 = append(sc.scores2, score)
		default:
			return nil, errors.New("channel number out of bound")
		}
	}

	addGroupAvgEmotion(nextDay-timeUnitInSec, group, groups)
	return groups, nil
}

func groupAvgDuration(_t1, _t2 uint64, appid string, filter string) (map[string][]*SimpleAvgDuration, error) {

	querySQL := fmt.Sprintf("select %s,%s,%s,%s from %s where %s=? and %s=1 and %s>=? and %s<=? order by %s\n",
		NID, NFILET, filter, NRDURATION, MainTable, NAPPID, NANARES, NFILET, NFILET, NFILET)

	rows, err := db.Query(querySQL, appid, _t1, _t2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t1 := int64(_t1)
	timeUnit := Day
	timeUnitInSec := int64(24 * 60 * 60)
	nextDay := AddTimeUnit(t1, timeUnit)

	var createdTime int64
	var lastID uint64
	var sc *ScoreCount
	var ok bool

	groupScore := make(map[string]*ScoreCount)
	groups := make(map[string][]*SimpleAvgDuration)

	for rows.Next() {
		var id uint64
		var tag sql.NullString
		var duration uint64

		err := rows.Scan(&id, &createdTime, &tag, &duration)

		if err != nil {
			return nil, err
		}

		if createdTime >= nextDay {

			addGroupAvgDuration(nextDay-timeUnitInSec, groupScore, groups)
			groupScore = make(map[string]*ScoreCount)
			for createdTime >= nextDay {
				nextDay = AddTimeUnit(nextDay, timeUnit)
			}

		}

		if lastID != id {
			lastID = id
			if sc, ok = groupScore[tag.String]; !ok {
				sc = &ScoreCount{}
				groupScore[tag.String] = sc
			}
		}

		dur := float64(duration) / 1000
		sc.count1++
		sc.score1 += dur
		sc.scores1 = append(sc.scores1, dur)
	}

	addGroupAvgDuration(nextDay-timeUnitInSec, groupScore, groups)
	return groups, nil
}

func dailyAvgEmotion(_t1 uint64, _t2 uint64, appid string, groupIDs []interface{}) ([]*AvgEmotion, error) {

	selectColumns := [2][]string{{NID, NRDURATION, NFILET}, {NCHANNEL, NSCORE}}
	whereStates, params := genDefaultWhereStates(_t1, _t2, appid)
	orderStates := genDefaultOrderStates()

	rows, err := getEmotionData(selectColumns, whereStates, groupIDs, orderStates, params)
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

	t2 := AddTimeUnit(int64(_t2), Day)

	avgEmotions := make([]*AvgEmotion, 0)

	for rows.Next() {
		var id, duration uint64
		var ch int8
		var score float64

		err := rows.Scan(&id, &duration, &createdTime, &ch, &score)

		if err != nil {
			return nil, err
		}

		for createdTime >= nextDay {
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

	for nextDay <= t2 {
		ae := genAvgEmotion(nextDay-1, totalCount, sumDuration, count1, score1, count2, score2)
		avgEmotions = append(avgEmotions, ae)
		nextDay = AddTimeUnit(nextDay, Day)
		totalCount, sumDuration = 0, 0
		count1, count2 = 0, 0
		score1, score2 = 0, 0
	}

	return avgEmotions, nil
}
