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

/*
//QueryEmotion for basePath/emotion/{serviceId}
func QueryEmotion(w http.ResponseWriter, r *http.Request) {

	appid := r.Header.Get(NUAPPID)

	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var res []byte
	paths := strings.SplitN(r.URL.Path, "/", MaxSlash)

	if !isFileID(paths[MaxSlash-1]) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var status int
	var err error
	fileID := paths[MaxSlash-1]
	res, status, err = QuerySingleResult(fileID, appid)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	contentType := "application/json; charset=utf-8"

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}
*/

func QueryEmotions(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	qas := new(QueryArgs)

	cursor, err := pieceCursor(qas, r)
	if err != nil {
		http.Error(w, "Bad request."+err.Error(), http.StatusBadRequest)
		return
	}

	res, status, err := Query(appid, cursor)

	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	contentType := "application/json; charset=utf-8"

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}

func QueryContinue(w http.ResponseWriter, r *http.Request) {

	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		var qcursor QCursor
		if r.Body == nil {
			http.Error(w, "No request body", http.StatusBadRequest)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&qcursor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cursor := DecryptCursor(qcursor.Cursor)
		if cursor == "" {
			http.Error(w, "Wrong cursor", http.StatusBadRequest)
			return
		}

		res, status, err := Query(appid, cursor)

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

func Query(appid string, cursor string) ([]byte, int, error) {
	conditions, conditions2, _, offset, remainCursor := makeCondition(cursor)
	var countInt int
	var err error

	if conditions == "" {
		return nil, http.StatusBadRequest, errors.New("Cursor parse error")
	}
	/*
		if count == "" {
			countInt, err = QueryCount(NAPPID + " = \"" + appid + "\" and " + conditions)
			if err != nil {
				log.Println(err)
				return nil, http.StatusInternalServerError, errors.New("Internal server error")
			}
		} else {
			countInt, err = strconv.Atoi(count)
			if err != nil {
				log.Printf(err.Error() + " count is not int")
				return nil, http.StatusBadRequest, errors.New("Cursor count error")
			}
		}
	*/

	rbs := make([]*ReturnBlock, 0)
	rp := new(ResultPage)

	conditions += " order by " + NFILET + " limit " + PAGELIMIT

	//parse offset
	var offsetInt int
	if offset != "" {
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			log.Printf(err.Error() + " offset is not int")
			return nil, http.StatusBadRequest, errors.New("Cursor offset error")
		}
		conditions += " offset " + offset
	}
	//rp.Offset = offsetInt

	//query result
	countInt, status, err := QueryResult(appid, conditions, conditions2, &rbs)
	if err != nil {
		return nil, status, err
	}

	//if countInt > offsetInt+PAGELIMITINT {
	if countInt >= PAGELIMITINT {
		//rp.HasMore = true

		//make new cursor
		//offsetInt += PAGELIMITINT
		offsetInt += countInt
		var newCursor string
		newCursor = strconv.Itoa(countInt)
		newCursor += "," + strconv.Itoa(offsetInt) + "," + remainCursor
		rp.Cursor = CreateCursor(newCursor)
	}

	rp.Total = countInt
	rp.Blocks = rbs

	encodeRes, err := json.Marshal(rp)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	return encodeRes, http.StatusOK, nil
}

func QueryEmotionDetail(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	paths := strings.SplitN(r.URL.Path, "/", MaxSlash)

	if !isFileID(paths[MaxSlash-1]) {
		http.Error(w, "Bad request: wrong file_id", http.StatusBadRequest)
		return
	}

	drb := new(DetailReturnBlock)

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

}

func GenerateReport(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	params := r.URL.Query()
	t1 := params.Get(NT1)
	t2 := params.Get(NT2)
	export := params.Get(NEXPORT)

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

	conditions := NUPT + ">=" + t1 + " and " + NUPT + "<=" + t2

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
}

func pieceCursor(qas *QueryArgs, r *http.Request) (string, error) {
	err := parseArgs(qas, r)
	if err != nil {
		return "", err
	}
	return pieceup(qas)
}

func parseArgs(qas *QueryArgs, r *http.Request) error {
	params := r.URL.Query()
	var err error

	qas.T1, err = parseTime(params.Get(NT1), true)
	if err != nil {
		return errors.New("t1 - " + err.Error())
	}

	qas.T2, err = parseTime(params.Get(NT2), false)
	if err != nil {
		return errors.New("t2 - " + err.Error())
	}

	if qas.T2 <= qas.T1 {
		return errors.New("t2<=t1")
	}

	qas.FileName = params.Get(NFILENAME)
	qas.Tag = params.Get(NTAG)
	qas.Tag2 = params.Get(NTAG2)

	qas.Status = params.Get(NSTATUS)

	if params.Get(NSCOREANG1) != "" {
		ch1Anger, err := strconv.ParseFloat(params.Get(NSCOREANG1), 64)
		if err != nil || ch1Anger < 0 || ch1Anger > 100 {
			return errors.New("Invalid ch1 anger score")
		}
	}
	qas.Ch1Anger = params.Get(NSCOREANG1)

	if params.Get(NSCOREANG2) != "" {
		ch2Anger, err := strconv.ParseFloat(params.Get(NSCOREANG2), 64)
		if err != nil || ch2Anger < 0 || ch2Anger > 100 {
			return errors.New("Invalid ch2 anger score")
		}
	}
	qas.Ch2Anger = params.Get(NSCOREANG2)

	return nil

}

func pieceup(qas *QueryArgs) (string, error) {
	var cursor string

	cursor = ",," //first two count,offset
	cursor += ">=" + qas.T1
	cursor += "," + "<=" + qas.T2

	cursor += ","
	if qas.FileName != "" {
		cursor += "=\"" + qas.FileName + "\""
	}

	cursor += ","
	if qas.Tag != "" {
		cursor += "=\"" + qas.Tag + "\""
	}

	cursor += ","
	if qas.Tag2 != "" {
		cursor += "=\"" + qas.Tag2 + "\""
	}

	switch qas.Status {
	case "done":
		cursor += ",>0,"
		if qas.Ch1Anger != "" {
			cursor += ">=" + qas.Ch1Anger
		}
		cursor += ","
		if qas.Ch2Anger != "" {
			cursor += ">=" + qas.Ch2Anger
		}
	case "":
		fallthrough
	case "all":
		cursor += ",,"
		if qas.Ch1Anger != "" {
			cursor += ">=" + qas.Ch1Anger
		}
		cursor += ","
		if qas.Ch2Anger != "" {
			cursor += ">=" + qas.Ch2Anger
		}
	case "wait":
		cursor += "," + "=-1" + ",,"
	default:
		return "", errors.New("Wrong status: " + qas.Status)
	}

	//	log.Println("cursor:" + cursor)
	//log.Println("encrypt cursor: " + CreateCursor(cursor))

	return cursor, nil

}

func parseTime(t string, isStart bool) (string, error) {
	now := time.Now()
	var tx string
	if t == "" {
		if isStart {
			days := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			tx = strconv.FormatInt(days.Unix(), 10)
		} else {
			tx = strconv.FormatInt(now.Unix(), 10)
		}

	} else {

		userT, err := strconv.ParseInt(t, 10, 64)

		if err != nil {
			return "", errors.New("wrong time format")
		}
		if userT < 0 {
			return "", errors.New("don't do early time")
		} else if userT > now.Unix() {
			return "", errors.New("don't do future time. Time traveler")

		}
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

//generate the real condition for database query,
//return condition, conditions2, count, offset and remaining cursor which excludes count and offset
func makeCondition(cursor string) (string, string, string, string, string) {
	cursorValues := strings.Split(cursor, ",")
	if len(cursorValues) != len(CursorFieldName) {
		log.Printf("splits: %q doesn't match cursor field\n", cursorValues)
		return "", "", "", "", ""
	}

	var conditions string
	var remainCursor string
	conditions = CursorFieldName[2] + cursorValues[2]
	remainCursor = cursorValues[2]
	for i := 3; i < IndexJoin; i++ {
		if cursorValues[i] != "" {
			conditions += " and " + CursorFieldName[i] + cursorValues[i]
		}
		remainCursor += "," + cursorValues[i]
	}

	var conditions2 string

	if cursorValues[8] != "" || cursorValues[9] != "" {
		conditions2 = "b." + NID + " in ( "
		var ch1AngerCondition string

		if cursorValues[8] != "" {
			ch1AngerCondition = "select " + NID + " from " + ChannelTable + " where " +
				NCHANNEL + "=1 and " + NEMOTYPE + "=0 and " + NSCORE + cursorValues[8]
		}
		if cursorValues[9] != "" {
			conditions2 += "select " + NID + " from " + ChannelTable + " where "
			if ch1AngerCondition != "" {
				conditions2 += NID + " in (" + ch1AngerCondition + ")" + " and "
			}
			conditions2 += NCHANNEL + "=2 and " + NEMOTYPE + "=0 and " + NSCORE + cursorValues[9]
		} else {
			conditions2 += ch1AngerCondition
		}

		conditions2 += ")"
		/*
			conditions2 = "b." + NEMOTIONTYPE + "=0 and ("
			if cursorValues[8] != "" {
				conditions2 += " (b." + NCHANNEL + "=1 and b." + NSCORE + cursorValues[8] + ")"
				if cursorValues[9] != "" {
					conditions2 += " or"
				}
			}
			if cursorValues[9] != "" {
				conditions2 += " (b." + NCHANNEL + "=2 and b." + NSCORE + cursorValues[9] + ")"
			}
			conditions2 += ")"
		*/
	}

	remainCursor += "," + cursorValues[8]
	remainCursor += "," + cursorValues[9]

	//log.Println(conditions)

	return conditions, conditions2, cursorValues[0], cursorValues[1], remainCursor
}
