package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		appid := r.Header.Get(HXAPPID)

		qas := new(QueryArgs)
		if r.Body != nil {
			err := json.NewDecoder(r.Body).Decode(qas)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		defer r.Body.Close()

		err := parseArgs(qas)
		if err != nil {
			http.Error(w, "Bad request."+err.Error(), http.StatusBadRequest)
			return
		}

		timeUnit, err := GetTimeUnit(qas)
		if err != nil {
			http.Error(w, "Bad request."+err.Error(), http.StatusBadRequest)
			return
		}

		s, status, err := QueryStat(qas, timeUnit, appid)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		res, err := json.Marshal(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		contentType := "application/json; charset=utf-8"

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(res)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}

//QueryStat query and do statistics
func QueryStat(qas *QueryArgs, timeUnit int64, appid string) (*Statistics, int, error) {

	conditions, params, err := makeCondition(qas)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	paramsWithAppid := make([]interface{}, 1, 16)
	paramsWithAppid[0] = appid
	paramsWithAppid = append(paramsWithAppid, params...)

	query := "with joinr as ("
	query += QueryFileInfo + " where " + NAPPID + "=?"

	if conditions != "" {
		query += conditions
	}

	query += " )"
	//query += "select " + NID + "," + NFILET + "," + NCHANNEL + "," + NEMOTYPE + "," + NSCORE + " from joinr order by " + NFILET + "," + NID

	query += "select " + NID + "," + NFILET + "," + NCHANNEL + "," + NEMOTYPE + "," + NSCORE + "," + NRDURATION + " from joinr group by " +
		NID + "," + NFILET + "," + NCHANNEL + "," + NEMOTYPE + "," + NSCORE + " order by " + NFILET + "," + NID

	//log.Println(query)
	//log.Println(paramsWithAppid)

	rows, err := db.Query(query, paramsWithAppid...)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	defer rows.Close()

	data, totalStat, err := getDataUnit(qas, timeUnit, rows)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	/*
		fmt.Println("---------------------------------")

		fmt.Printf("count:%v\n", totalStat.count)
		fmt.Printf("duration:%v\n", totalStat.duration)
		fmt.Printf("ch1_score:%v,ch1_count:%v\n", totalStat.ch1Score, totalStat.ch1Count)
		fmt.Printf("ch2_score:%v,ch2_count:%v\n", totalStat.ch2Score, totalStat.ch2Count)
		fmt.Printf("data len:%d\n", len(data))
		fmt.Println("---------------------------------")
	*/

	s := new(Statistics)
	s.Count = totalStat.count
	s.Duration = totalStat.duration / 1000
	if totalStat.ch1Count != 0 {
		s.Ch1AvgAnger = float64(int64((totalStat.ch1Score/float64(totalStat.ch1Count))*100)) / 100
	}
	if totalStat.ch2Count != 0 {
		s.Ch2AvgAnger = float64(int64((totalStat.ch2Score/float64(totalStat.ch2Count))*100)) / 100
	}

	s.Data = data
	s.From = qas.T1
	s.To = qas.T2
	s.TimeUnit = TIMEUNITNAME[timeUnit]

	return s, http.StatusOK, nil
}

//time unit
const (
	Hour  = int64(time.Hour / time.Second)
	Day   = Hour * 24
	Week  = Day * 7
	Month = Day * 30
	Year  = Day * 365
)

var TIMEUNITNAME = map[int64]string{Hour: "hour", Day: "day", Month: "month", Year: "year"}

var TIMEUNIT = [...]int64{Hour, Day, Month, Year}

func GetTimeUnit(qas *QueryArgs) (int64, error) {
	diff := qas.T2 - qas.T1
	if diff <= 0 {
		return 0, errors.New("t2-t1<=0")
	}

	timeUnit := TIMEUNIT[0]
	for i := len(TIMEUNIT) - 1; i >= 0; i-- {
		if diff > TIMEUNIT[i] {
			timeUnit = TIMEUNIT[i]
			break
		}
	}

	return timeUnit, nil
}

func RoundUpTime(t int64, unit int64) (int64, error) {

	t1 := time.Unix(t, 0)
	switch unit {
	case Hour:
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), 0, 0, 0, t1.Location())
	case Day:
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	case Month:
		t1 = time.Date(t1.Year(), t1.Month(), 1, 0, 0, 0, 0, t1.Location())
	case Year:
		t1 = time.Date(t1.Year(), 1, 1, 0, 0, 0, 0, t1.Location())
	default:
		return 0, errors.New("error time duration:" + strconv.FormatInt(unit, 10))
	}

	return t1.Unix(), nil
}

func addTimeUnit(t int64, unit int64) int64 {

	t1 := time.Unix(t, 0)

	switch unit {
	case Hour:
		return t1.Unix() + Hour
	case Day:
		t1 = t1.AddDate(0, 0, 1)
	case Month:
		t1 = t1.AddDate(0, 1, 0)
	case Year:
		t1 = t1.AddDate(1, 0, 0)
	default:
		return 0
	}
	return t1.Unix()
}

type statCounting struct {
	ch1Score       float64
	ch2Score       float64
	count          uint64
	ch1Count       uint64
	ch2Count       uint64
	duration       uint64
	startTimeStamp int64
	nextTimeStamp  int64
	timeUnit       int64
	maxTimeStamp   int64
	finished       bool
}

func getDataUnit(qas *QueryArgs, timeUnit int64, rows *sql.Rows) ([]*StatUnit, *statCounting, error) {
	sus := make([]*StatUnit, 0)
	var sc statCounting
	var err error
	var lastID int64

	totalSc := &statCounting{}

	//compute the timestamp of next time
	sc.nextTimeStamp, err = RoundUpTime(qas.T1, timeUnit)
	if err != nil {
		return nil, nil, err
	}
	sc.nextTimeStamp = addTimeUnit(sc.nextTimeStamp, timeUnit)

	sc.startTimeStamp = qas.T1
	sc.timeUnit = timeUnit
	sc.maxTimeStamp = qas.T2
	/*
		log.Printf("%s(%v)\n", time.Unix(qas.T1, 0), qas.T1)
		log.Printf("%s(%v)\n", time.Unix(qas.T2, 0), qas.T2)
		log.Println(time.Unix(sc.nextTimeStamp, 0))
		log.Println("----------------------------")
	*/
	for {

		var id, createdT int64
		var channel sql.NullInt64
		var emotionType sql.NullInt64
		var score sql.NullFloat64
		var duration sql.NullInt64

		if rows.Next() {
			err := rows.Scan(&id, &createdT, &channel, &emotionType, &score, &duration)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
		} else {

			if sc.count > 0 {
				su := makeStatUnit(&sc)
				sus = append(sus, su)
			}

			//fill up the empty duration
			for sc.nextTimeStamp <= qas.T2 {
				su := makeStatUnit(&sc)
				sus = append(sus, su)
			}

			//fill the last duration
			if !sc.finished {
				su := makeStatUnit(&sc)
				sus = append(sus, su)
			}

			break
		}

		for createdT >= sc.nextTimeStamp {
			su := makeStatUnit(&sc)
			sus = append(sus, su)
		}

		if lastID != id {
			sc.count++
			totalSc.count++
			lastID = id

			if duration.Valid && duration.Int64 != -1 {
				sc.duration += uint64(duration.Int64)
				totalSc.duration += uint64(duration.Int64)
			}

		}

		if channel.Valid && score.Valid && emotionType.Valid && emotionType.Int64 == NANGER {
			if channel.Int64 == 1 {
				sc.ch1Score += score.Float64
				sc.ch1Count++
				totalSc.ch1Score += score.Float64
				totalSc.ch1Count++

			} else if channel.Int64 == 2 {
				sc.ch2Score += score.Float64
				sc.ch2Count++
				totalSc.ch2Score += score.Float64
				totalSc.ch2Count++
			}

		}
	}

	/*
		for idx, su := range sus {
			log.Printf("%d. %s~%s count:%d, av1:%lf, av2:%lf\n", idx+1, time.Unix(su.From, 0), time.Unix(su.To, 0), su.Count, su.Ch1AvgAnger, su.Ch2AvgAnger)
		}
	*/

	return sus, totalSc, nil
}

func makeStatUnit(sc *statCounting) *StatUnit {
	su := new(StatUnit)
	su.From = sc.startTimeStamp
	su.To = sc.nextTimeStamp - 1
	if sc.ch1Count != 0 {
		su.Ch1AvgAnger = float64(int64((sc.ch1Score/float64(sc.ch1Count))*100)) / 100
	}
	if sc.ch2Count != 0 {
		su.Ch2AvgAnger = float64(int64((sc.ch2Score/float64(sc.ch2Count))*100)) / 100
	}

	su.Duration = sc.duration / 1000
	su.Count = sc.count
	//sus = append(sus, su)

	sc.ch1Count = 0
	sc.ch2Count = 0
	sc.ch1Score = 0
	sc.ch2Score = 0
	sc.count = 0
	sc.duration = 0
	sc.startTimeStamp = sc.nextTimeStamp

	sc.nextTimeStamp = addTimeUnit(sc.nextTimeStamp, sc.timeUnit)

	if su.To >= sc.maxTimeStamp {
		su.To = sc.maxTimeStamp
		sc.finished = true
	}
	return su
}

/*
func getDataUnit(qas *QueryArgs, timeUnit int64, rows *sql.Rows) ([]*StatUnit, error) {

	sus := make([]*StatUnit, 0)
	var s int64
	var lastID int64
	var count int64

	numOfEmotion := len(DefaultEmotion)
	criteria := make([]int, len(DefaultEmotion))
	hasEmotion := make([]bool, numOfEmotion)
	emotionCount := make([]int64, numOfEmotion)

	//setting score crtieria slice
	for _, v := range DefaultRevertEmotion {
		score, _ := strconv.Atoi(v[1])
		idx, _ := strconv.Atoi(v[0])
		criteria[idx] = score
	}

	//compute the timestamp of next time
	nextTimeStamp, err := roundUpTime(qas.T1, timeUnit)
	if err != nil {
		return nil, err
	}
	nextTimeStamp = addTimeUnit(nextTimeStamp, timeUnit)

	s = qas.T1

	for rows.Next() {

		var id, createdT int64
		var channel sql.NullInt64
		var emotionType sql.NullInt64
		var score sql.NullFloat64

		err := rows.Scan(&id, &createdT, &channel, &emotionType, &score)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		for createdT >= nextTimeStamp {

			su := new(StatUnit)

			su.Total = count
			su.From = s
			su.To = nextTimeStamp - 1
			su.NeutralCount = emotionCount[0]
			su.AngerCount = emotionCount[1]

			sus = append(sus, su)

			count = 0
			s = nextTimeStamp

			for i := 0; i < numOfEmotion; i++ {
				hasEmotion[i] = false
			}
			emotionCount[0] = 0
			emotionCount[1] = 0
			nextTimeStamp = addTimeUnit(nextTimeStamp, timeUnit)

		}

		if id != lastID {
			count++

			for i := 0; i < numOfEmotion; i++ {
				hasEmotion[i] = false
			}

			lastID = id
		}

		if emotionType.Valid && score.Valid {
			idx := emotionType.Int64
			if !hasEmotion[idx] && int(score.Float64) >= criteria[idx] {
				hasEmotion[idx] = true
				emotionCount[idx]++
			}
		}

	}

	if count > 0 {

		//pading empty duration
		for qas.T2-s > timeUnit {
			su := new(StatUnit)
			su.From = s
			roundUpTime, _ := roundUpTime(s, timeUnit)

			s = addTimeUnit(roundUpTime, timeUnit)
			su.To = s - 1

			sus = append(sus, su)
		}

		su := new(StatUnit)

		su.Total = count
		su.From = s
		su.To = qas.T2
		su.NeutralCount = emotionCount[0]
		su.AngerCount = emotionCount[1]

		sus = append(sus, su)
	}

	return sus, nil
}
*/
