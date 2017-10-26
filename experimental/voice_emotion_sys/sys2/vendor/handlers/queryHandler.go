package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var isFileID = regexp.MustCompile("^[a-f0-9]+$").MatchString

const MaxSlash = 6

func QueryEmotions(w http.ResponseWriter, r *http.Request) {

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
		cas := new(CursorArgs)
		cas.Qas = qas

		res, status, err := Query(appid, cas)

		if err != nil {
			http.Error(w, err.Error(), status)
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

func QueryContinue(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		var qcursor QCursor
		appid := r.Header.Get(HXAPPID)
		if r.Body == nil {
			http.Error(w, "No request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&qcursor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, err := DecryptCursor(qcursor.Cursor)
		if err != nil {
			log.Println(err)
			http.Error(w, "Wrong cursor", http.StatusBadRequest)
			return
		}

		cas := new(CursorArgs)

		err = json.Unmarshal(b, cas)
		if err != nil {
			http.Error(w, "Wrong cursor", http.StatusBadRequest)
			return
		}

		res, status, err := Query(appid, cas)
		if err != nil {
			http.Error(w, err.Error(), status)
			log.Printf("[%s]\n", qcursor.Cursor)
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

func Query(appid string, cas *CursorArgs) ([]byte, int, error) {

	conditions, params, err := makeCondition(cas.Qas)

	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	paramsWithAppid := make([]interface{}, 1, 16)
	paramsWithAppid[0] = appid
	paramsWithAppid = append(paramsWithAppid, params...)

	if cas.Count == 0 {
		cas.Count, err = QueryCount(conditions, paramsWithAppid...)
		if err != nil {
			log.Println(err)
			return nil, http.StatusInternalServerError, errors.New("Internal server error")
		}
	}

	rp := new(ResultPage)

	rbs, status, err := QueryResult(cas.Offset, conditions, paramsWithAppid...)
	if err != nil {
		return nil, status, errors.New("Internal server error")
	}

	//update offset and make new cursor
	if cas.Count > cas.Offset+PAGELIMITINT {
		cas.Offset += PAGELIMITINT
		b, _ := json.Marshal(cas)
		rp.Cursor = CreateCursor(b)
	}

	rp.Total = cas.Count
	rp.Blocks = rbs

	if cas.Count != 0 && len(rbs) == 0 {
		log.Printf("[Warning] %s took other's cursor to query\n", appid)
		return nil, http.StatusBadRequest, errors.New("Bad request")
	}

	encodeRes, err := json.Marshal(rp)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	return encodeRes, http.StatusOK, nil
}

func QueryEmotionDetail(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		appid := r.Header.Get(HXAPPID)
		paths := strings.SplitN(r.URL.Path, "/", MaxSlash)

		if !isFileID(paths[MaxSlash-1]) {
			http.Error(w, "Bad request: wrong file_id", http.StatusBadRequest)
			return
		}

		drb := new(DetailReturnBlock)
		drb.Tags = make([]string, 0, LIMITUSERTAGS)

		status, err := QuerySingleDetail(paths[MaxSlash-1], appid, drb)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		encodeRes, err := json.Marshal(drb)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		contentType := ContentTypeJSON

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(encodeRes)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}

func GenerateReport(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		appid := r.Header.Get(HXAPPID)
		if appid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		now := time.Now()
		thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		params := r.URL.Query()
		t1 := params.Get(QT1)
		t2 := params.Get(QT2)
		export := params.Get(QEXPORT)

		if t1 == "" {
			t1 = strconv.FormatInt(thisMonth.Unix(), 10)
		}
		if t2 == "" {
			t2 = strconv.FormatInt(now.Unix(), 10)
		}

		t1UTC, _ := strconv.ParseInt(t1, 10, 64)
		t2UTC, _ := strconv.ParseInt(t2, 10, 64)

		if t2UTC > now.Unix() || t1UTC > now.Unix() {
			http.Error(w, "Don't do future time. Time traveler", http.StatusBadRequest)
			return
		}

		if t1UTC >= t2UTC {
			http.Error(w, "t1>=t2", http.StatusBadRequest)
			return
		}

		if (t2UTC - t1UTC) > ReportLimitDay {
			http.Error(w, "Only support query range < 90 days", http.StatusBadRequest)
			return
		}

		conditions := NFILET + ">=" + t1 + " and " + NFILET + "<=" + t2

		//fileFullPath := "upload_file/" + appid + "/" + randomString(32) + "-report." + export

		rc := new(ReportCtx)

		err := rc.InitReportCtx(export, "", false)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer rc.CloseCtx()

		reportPeriod := make(map[string]int64)
		reportPeriod[NFROM] = t1UTC
		reportPeriod[NTO] = t2UTC

		rc.PutHeader(reportPeriod)

		QueryReport(appid, conditions, rc)

		b, contentType, err := rc.FinishReport()
		if err != nil {
			http.Error(w, "Internal serval error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(b)
		/*
			_, _, err = rc.FinishReport()
			if err != nil {
				http.Error(w, "Internal serval error", http.StatusInternalServerError)
				return
			}
			http.ServeFile(w, r, fileFullPath)
		*/

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}

func parseArgs(qas *QueryArgs) error {
	var err error

	//check time
	qas.T1, err = parseTime(qas.T1, true)
	if err != nil {
		return errors.New("t1 - " + err.Error())
	}

	qas.T2, err = parseTime(qas.T2, false)
	if err != nil {
		return errors.New("t2 - " + err.Error())
	}

	if qas.T2 <= qas.T1 {
		return errors.New(QT2 + "<=" + QT1)
	}

	//check tags uniquely
	if hasDupTags(qas.Tags) {
		return errors.New("has duplicate tags")
	}

	return nil

}

func parseTime(t int64, isStart bool) (int64, error) {
	now := time.Now()
	var tx int64
	if t == 0 {
		if isStart {
			days := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			tx = days.Unix()
		} else {
			tx = now.Unix()
		}

	} else {

		userT := t
		if userT < 0 {
			return 0, errors.New("don't do early time")
		}
		/*
			else if userT > now.Unix() {
				return 0, errors.New("don't do future time. Time traveler")

			}
		*/
		/*
			t += "+08"
			txT, err := time.Parse("2006010215-07", t)
			if err != nil {
				return "", err
			}
		*/
		tx = t
	}
	return tx, nil
}

//return condition, parameters,error message
func makeCondition(qas *QueryArgs) (string, []interface{}, error) {
	var conditions string
	params := make([]interface{}, 0, 16)
	if qas.T1 != 0 {
		conditions += " and " + NFILET + ">=?"
		params = append(params, qas.T1)
	}
	if qas.T2 != 0 {
		conditions += " and " + NFILET + "<=?"
		params = append(params, qas.T2)
	}
	if qas.FileName != "" {
		conditions += " and " + NFILENAME + " like concat('%',?,'%')"
		params = append(params, qas.FileName)
	}

	doScoreCondition := true

	switch qas.Status {
	case "wait":
		doScoreCondition = false
		conditions += " and " + NANARES + " =-1"
	case "done":
		conditions += " and " + NANARES + " >0"
	case "":
	case "all":
	default:
		return "", nil, errors.New("wrong status:" + qas.Status)
	}

	/*
		if doScoreCondition &&
			(len(qas.Ch1Emotion) > 0 || len(qas.Ch2Emotion) > 0) {
			count := 0

			conditions += " and a." + NID + " in (select " + NID + " from " + ChannelTable + " where "
			err := makeEmotionCondition(qas.Ch1Emotion, "1", &count, &conditions, &params)
			if err != nil {
				return "", nil, err
			}
			err = makeEmotionCondition(qas.Ch2Emotion, "2", &count, &conditions, &params)
			if err != nil {
				return "", nil, err
			}

			conditions += " group by " + NID + " having count(*)=" + strconv.Itoa(count) + ")"
		}
	*/

	if doScoreCondition && (qas.Ch1Anger > 0 || qas.Ch2Anger > 0) {
		count := 0
		conditions += " and a." + NID + " in (select " + NID + " from " + ChannelTable + " where "

		if qas.Ch1Anger > 0 {
			makeAngerCondition("1", &count, &conditions)
			params = append(params, qas.Ch1Anger)
		}
		if qas.Ch2Anger > 0 {
			makeAngerCondition("2", &count, &conditions)
			params = append(params, qas.Ch2Anger)
		}

		conditions += " group by " + NID + " having count(*)=" + strconv.Itoa(count) + ")"
	}

	if len(qas.Tags) > 0 {
		count := 0
		conditions += " and a." + NID + " in (select " + NID + " from " + UserDefinedTagsTable + " where "
		err := makeTagsCondition(qas.Tags, &count, &conditions, &params)
		if err != nil {
			return "", nil, err
		}
		conditions += " group by " + NID + " having count(*)=" + strconv.Itoa(count) + ")"
	}

	if len(qas.UsrColumn) > 0 {
		count := 0
		conditions += " and a." + NID + " in (select " + NID + " from " + UsrColValTable + " where "
		err := makeUsrColumnCondition(qas.UsrColumn, &count, &conditions, &params)
		if err != nil {
			return "", nil, err
		}
		conditions += " group by " + NID + " having count(*)=" + strconv.Itoa(count) + ")"
	}
	return conditions, params, nil

}

func makeUsrColumnCondition(cvs []*ColumnValue, count *int, conditions *string, params *[]interface{}) error {
	for _, cv := range cvs {
		if *count != 0 {
			*conditions += " or "
		}
		*conditions += "(" + NCOLID + "=? and " + NCOLVAL + "=? )"
		*params = append(*params, cv.ColID)
		*params = append(*params, cv.Value)
		*count++
	}
	return nil
}

func makeTagsCondition(tags []string, count *int, conditions *string, params *[]interface{}) error {
	for _, tag := range tags {
		if *count != 0 {
			*conditions += " or "
		}
		*conditions += NTAG + "=?"
		*params = append(*params, tag)
		*count++
	}
	return nil
}

func makeAngerCondition(channel string, count *int, conditions *string) {
	if *count != 0 {
		*conditions += " or"
	}

	*conditions += " (" + NCHANNEL + "=" + channel + " and " + NEMOTYPE + "=" + DefaultRevertEmotion[EANGER][0] + " and " + NSCORE + " >=?)"
	*count++
}

func makeEmotionCondition(emotions []string, channel string, count *int, conditions *string, params *[]interface{}) error {
	for _, v := range emotions {
		if *count != 0 {
			*conditions += " or"
		}
		emotion, ok := DefaultRevertEmotion[v]
		if !ok {
			return errors.New("Wrong emotion:" + v)
		}
		*conditions += " (" + NCHANNEL + "=" + channel + " and " + NEMOTYPE + "=" + emotion[0] + " and " + NSCORE + " >=?)"
		*params = append(*params, emotion[1])
		*count++
	}
	return nil

}
