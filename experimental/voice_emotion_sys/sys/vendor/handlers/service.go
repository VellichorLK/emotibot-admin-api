package handlers

import (
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

func avgEmotionByTag(_t1 uint64, _t2 uint64, appid string, filter string) ([]*AvgEmotion, error) {

	selectColumns := [2][]string{{NID, NFILET, filter}, {NCHANNEL, NSCORE}}
	whereStates, params := genDefaultWhereStates(_t1, _t2, appid)
	orderStates := genDefaultOrderStates()

	rows, err := getEmotionData(selectColumns, whereStates, orderStates, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_ = rows
	return nil, nil
}

func dailyAvgEmotion(_t1 uint64, _t2 uint64, appid string) ([]*AvgEmotion, int, error) {

	selectColumns := [2][]string{{NID, NRDURATION, NFILET}, {NCHANNEL, NSCORE}}
	whereStates, params := genDefaultWhereStates(_t1, _t2, appid)
	orderStates := genDefaultOrderStates()

	rows, err := getEmotionData(selectColumns, whereStates, orderStates, params)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}

	fmt.Println(params)

	t1 := int64(_t1)
	nextDay := AddTimeUnit(t1, Day)
	var totalCount uint64      //total audio file in one day
	var score1, score2 float64 //anger score of channel1 and channel2
	var count1, count2 uint64  //count of channel1 and channel2
	var sumDuration uint64
	var createdTime int64
	var lastID uint64
	var numOfUnits int

	avgEmotions := make([]*AvgEmotion, 0)

	for rows.Next() {
		var id, duration uint64
		var ch int8
		var score float64

		err := rows.Scan(&id, &duration, &createdTime, &ch, &score)

		if err != nil {
			return nil, 0, err
		}

		if createdTime >= nextDay {
			ae := genAvgEmotion(nextDay-1, totalCount, sumDuration, count1, score1, count2, score2)
			fmt.Println(ae)
			avgEmotions = append(avgEmotions, ae)

			totalCount, sumDuration = 0, 0
			count1, count2 = 0, 0
			score1, score2 = 0, 0

			numOfUnits++
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
			return nil, 9, errors.New("channel number out of bound")
		}

	}

	numOfUnits++
	ae := genAvgEmotion(nextDay-1, totalCount, sumDuration, count1, score1, count2, score2)
	avgEmotions = append(avgEmotions, ae)

	return avgEmotions, numOfUnits, nil
}

func parseAvgEmotions(_t1 uint64, _t2 uint64, unit int64, valuePtr []interface{}) {

}
