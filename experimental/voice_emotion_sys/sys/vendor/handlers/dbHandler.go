package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"sort"

	"errors"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

//InitDatabaseCon create the initial connection to database
func InitDatabaseCon(ip string, port string, username string, password string, database string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if db == nil {
		src := username + ":" + password + "@tcp(" + ip + ":" + port + ")/" + database

		var err error
		db, err = sql.Open("mysql", src)
		if err != nil {
			log.Println(err)
			return err
		}
		err = db.Ping()
		if err != nil {
			log.Println(err)
			db.Close()
			db = nil
			return err
		}
	}

	return nil
}

//Close the db connection
func Close() {
	if db != nil {
		db.Close()
		db = nil
	}
}

//InsertFileInfoSQL used for insert a new file info record
const InsertFileInfoSQL = "insert into " + MainTable + " (" + NFILEID + ", " + NFILENAME + ", " + NFILETYPE + ", " +
	NDURATION + ", " + NFILET + ", " + NCHECKSUM + ", " + NTAG + ", " + NPRIORITY + ", " + NAPPID + ", " + NSIZE + ", " +
	NFILEPATH + ", " + NUPT + ", " + NTAG2 + ", " + NRDURATION + ")" + " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

const DeleteFileRowSQL = "delete from " + MainTable

//InsertAnalysisSQL used for insert a new emotion record
const InsertAnalysisSQL = "insert into " + AnalysisTable + " (" + NID + ", " + NSEGST + ", " + NSEGET + ", " +
	NCHANNEL + ", " + NSTATUS + ", " + NEXTAINFO + " )" + " values(?,?,?,?,?,?)"

//InsertEmotionSQL used for insert a new emotion record
const InsertEmotionSQL = "insert into " + EmotionTable + " (" + NSEGID + ", " + NEMOTYPE + ", " + NSCORE + " )" +
	" values(?,?,?)"

//QueryChannelScoreSQL used for query
const SingleQuerySQL = "select " + NFILEID + ", " + NFILENAME + ", " + NFILETYPE + ", " + NDURATION +
	", " + NFILET + ", " + NCHECKSUM + ", " + NTAG + ", " + NPRIORITY + ", " + NSIZE + " from " + MainTable
const QueryChannelScoreSQL = "select " + NCHANNEL + ", " + NEMOTYPE + ", " + NSCORE + " from " + ChannelTable
const QueryAnalysisSQL = "select extra_info from " + AnalysisTable

//QueryEmotionMapSQL query the emotion map table, only do once at the statup
const QueryEmotionMapSQL = "select " + NEMOTYPE + ", " + NEMOTION + " from " + EmotionMapTable

const InsertChannelScoreSQL = "insert into " + ChannelTable + " (" + NID + ", " + NCHANNEL + ", " + NEMOTYPE + ", " + NSCORE + ") " +
	" values(?,?,?,?)"

const UpdateResultSQL = "update " + MainTable + " set " + NANAST + "=? ," + NANAET + "=? ," + NANARES + "=? ," + NRDURATION +
	"=? where " + NID + "=?"

const QueryFileInfoAndChanScoreSQL = "select a." + NFILEID + ", a." + NFILENAME + ", a." + NFILETYPE + ", a." + NRDURATION +
	", a." + NFILET + ", a." + NCHECKSUM + ", a." + NTAG + ", a." + NTAG2 + ", a." + NPRIORITY + ", a." + NSIZE + ", a." + NANARES + ", a." + NUPT +
	", b." + NCHANNEL + ", b." + NEMOTYPE + ", b." + NSCORE

const QueryFileInfoAndChanScoreSQL2 = " as a left join " + ChannelTable + " as b on a." + NID + "=b." + NID

const QueryDetailSQL = "select * from (select " + NID + "," + NFILEID + "," + NFILENAME + "," + NFILETYPE + "," + NRDURATION + "," +
	NFILET + "," + NCHECKSUM + "," + NTAG + "," + NTAG2 + "," + NPRIORITY + "," + NSIZE + "," + NANARES + "," + NUPT +
	" from " + MainTable + ") as a left join ( select b." + NID + ",b." + NSEGST + ",b." + NSEGET +
	",b." + NCHANNEL + ",b." + NSTATUS + ",b." + NEXTAINFO + ",c." + NEMOTYPE + ",c." + NSCORE + " from ( select * from " +
	AnalysisTable + " where " + NID + "=(select " + NID + " from " + MainTable + " where " +
	NAPPID + "=? and " + NFILEID + "=?)) as b left join " + EmotionTable + " as c " +
	"on b." + NSEGID + "=c." + NSEGID + ") as d on a." + NID + "=d." + NID + " where a." + NFILEID + "=?"

const QueryReportSQL = "select " + NFILENAME + ", " + NRDURATION + "," + NTAG + ", " + NTAG2 + ", " + NUPT + ", " + NANAST + "," + NANAET + " from " + MainTable

func InsertFileRecord(fi *FileInfo) error {

	stmt, err := db.Prepare(InsertFileInfoSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(fi.FileID, fi.FileName, fi.FileType, fi.Duration, fi.CreateTime,
		fi.Checksum, fi.Tag, fi.Priority, fi.Appid, fi.Size, fi.Path, fi.UploadTime, fi.Tag2, 0)
	if err != nil {
		log.Println(err)
		return err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return err
	}

	fi.ID = strconv.FormatUint(uint64(lastID), 10)

	return nil
}

func DeleteFileRecord(id uint64) error {
	query := DeleteFileRowSQL
	query += " where " + NID + "=?"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

func InsertAnalysisRecord(eb *EmotionBlock) error {
	stmt, err := db.Prepare(InsertAnalysisSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	for _, d := range eb.Segments {

		enc, err := json.Marshal(d.ExtraInfo)
		if err != nil {
			log.Println(err)
			return err
		}

		res, err := stmt.Exec(eb.IDUint64, d.SegStartTime, d.SegEndTime, d.Channel, d.Status, enc)
		if err != nil {
			log.Println(err)
			return err
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			return err
		}
		for _, s := range d.ScoreList {
			label, ok := s.Label.(float64)
			if !ok {
				log.Printf("label is not float, %v %T\n", s.Label, s.Label)
			} else {

				err := InsertEmotionRecord(lastID, (int)(label), s.Score)
				if err != nil {
					return err
				}
			}

		}
	}

	return err
}

func ComputeChannelScore(eb *EmotionBlock) {

	const divider = 1000
	NumSegInOneChannel := make(map[int]int)
	TotalProbabilityWithEmotion := make(map[int]float64)
	NumOfHasOneEmotionInOneChannel := make(map[int]int)

	angerScore := make([]float64, 0)

	threashold := 0.7
	weight := 0.5

	for _, seg := range eb.Segments {
		for _, s := range seg.ScoreList {
			label, ok := s.Label.(float64)
			if !ok {
				log.Printf("label is not float, %v %T\n", s.Label, s.Label)
			} else {

				emotionChannelKey := seg.Channel*divider + int(label)
				NumSegInOneChannel[emotionChannelKey]++
				TotalProbabilityWithEmotion[emotionChannelKey] += s.Score
				if s.Score >= threashold {
					NumOfHasOneEmotionInOneChannel[emotionChannelKey]++
				}

				//if it's anger
				if int(label) == 1 && seg.Channel == 1 {
					angerScore = append(angerScore, s.Score)
				}
			}

		}
	}

	for chEmotion, count := range NumSegInOneChannel {
		var twoFixedScore, totalScore, weightScore, rateScore float64

		//log.Printf("id:%v, ch:%d, emotion:%d, score:%v, chemotion:%v\n", eb.IDUint64, chEmotion/divider, chEmotion%divider, totalScore*100, chEmotion)

		channel := chEmotion / divider
		emotionType := chEmotion % divider

		if count != 0 {

			if emotionType == 1 && channel == 1 {
				sort.Float64s(angerScore)
				var topCount int
				var length int
				var top5Score float64
				length = len(angerScore)
				if length > 5 {
					topCount = 5
				} else {
					topCount = length
				}

				for i := 0; i < topCount; i++ {
					top5Score += angerScore[length-1-i]
				}

				twoFixedScore = float64(int((top5Score/float64(topCount))*100)) / 100

			} else {
				weightScore = TotalProbabilityWithEmotion[chEmotion] / float64(count)
				rateScore = float64(NumOfHasOneEmotionInOneChannel[chEmotion]) / float64(count)
				totalScore = weightScore*weight + rateScore*(1-weight)
				twoFixedScore = float64(int(totalScore*100*100)) / 100
			}

			InsertChannelScore(eb.IDUint64, channel, emotionType, twoFixedScore)
		}
	}

	/*
		const divider = 100

		totalEmotionCount := make(map[int]int)
		emotionCount := make(map[int]int)
		for _, seg := range eb.Segments {
			for _, s := range seg.ScoreList {
				label, ok := s.Label.(float64)
				if !ok {
					log.Printf("label is not float, %v %T\n", s.Label, s.Label)
				} else {
					key := seg.Channel*divider + int(label)
					_, ok := emotionCount[key]
					if !ok {
						emotionCount[key] = 0
						totalEmotionCount[key] = 0
					}
					if s.Score > threashold {
						emotionCount[key]++
					}
					totalEmotionCount[key]++
				}
			}
		}

		for k, v := range emotionCount {
			channel := k / divider
			label := k % divider
			InsertChannelScore(eb.IDUint64, channel, label, float64(v)/float64(totalEmotionCount[k]))

			//log.Printf("chan: %v label:%v count:%v total:%v, score:%v\n", channel, label, v, totalEmotionCount[k], float64(v)/float64(totalEmotionCount[k]))
		}
	*/
}

func InsertChannelScore(id uint64, channel int, emotion int, score float64) error {

	stmt, err := db.Prepare(InsertChannelScoreSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, channel, emotion, score)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func UpdateResult(eb *EmotionBlock) error {
	stmt, err := db.Prepare(UpdateResultSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(eb.AnalysisStart, eb.AnalysisEnd, eb.Result, eb.RDuration, eb.IDUint64)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func QueryBlob() {

	rows, err := db.Query(QueryAnalysisSQL)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var blob []uint8

		err := rows.Scan(&blob)
		if err != nil {
			log.Println(err)
		}
		/*
			switch v := blob.(type) {
			case int:
				fmt.Printf("I don't know about type %T!\n", v)
			default:
				fmt.Printf("I don't know about type %T!\n", v)

			}
		*/

		s := string(blob[:])

		log.Println(s)
	}

}

func InsertEmotionRecord(emotionID int64, emotionType int, score float64) error {
	stmt, err := db.Prepare(InsertEmotionSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(emotionID, emotionType, score)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

/*
func QueryUnfinished(fileID string, appid string) ([]byte, int, error) {
	rb := new(ReturnBlock)
	rb.Channels = make([]*ChannelResult, 0)

	query := SingleQuerySQL
	query += " where " + NFILEID + "=? and " + NAPPID + "=?" + " and " + NANARES + "=-1"

	err := db.QueryRow(query, fileID, appid).Scan(&rb.FileID, &rb.FileName, &rb.FileType, &rb.Duration,
		&rb.CreateTime, &rb.Checksum, &rb.Tag, &rb.Priority, &rb.Size)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.StatusBadRequest, errors.New("No " + NFILEID + ": " + fileID)
		}
		log.Println(err)
		return nil, http.StatusInternalServerError, err
	}
	//encode to json
	encodeRes, err := json.Marshal(rb)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	return encodeRes, http.StatusOK, nil
}
*/
func QuerySingleDetail(fileID string, appid string, drb *DetailReturnBlock) (int, error) {
	//debug.FreeOSMemory()
	drb.Channels = make([]*DetailChannelResult, 0)

	query := QueryDetailSQL

	rows, err := db.Query(query, appid, fileID, fileID)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, errors.New("Internal server error")
	}
	defer rows.Close()

	channelMapping := make(map[int]*DetailChannelResult)
	vadMapping := make(map[string]*VadResult)
	count := 0
	var id uint64

	for rows.Next() {
		count++
		var ch, status, emType sql.NullInt64
		var st, ed, score sql.NullFloat64
		var file sql.NullInt64
		var blob []uint8

		err := rows.Scan(&id, &drb.FileID, &drb.FileName, &drb.FileType, &drb.Duration,
			&drb.CreateTime, &drb.Checksum, &drb.Tag, &drb.Tag2, &drb.Priority, &drb.Size, &drb.AnalysisResult, &drb.UploadTime,
			&file, &st, &ed, &ch, &status, &blob, &emType, &score)

		if err != nil {
			log.Println(err)
			return http.StatusInternalServerError, errors.New("Internal server error")
		}

		drb.Duration /= 1000

		if file.Valid && st.Valid && ed.Valid && ch.Valid && status.Valid {
			chInt := int(ch.Int64)
			dcr, ok := channelMapping[chInt]
			if !ok {
				dcr = new(DetailChannelResult)
				dcr.VadResults = make([]*VadResult, 0)
				dcr.Result = make([]*EmtionScore, 0)
				channelMapping[chInt] = dcr
				dcr.ChannelID = chInt
				drb.Channels = append(drb.Channels, dcr)
			}

			vadKey := strconv.Itoa(chInt) + strconv.FormatFloat(st.Float64, 'E', -1, 64) + strconv.FormatFloat(ed.Float64, 'E', -1, 64)

			vr, ok := vadMapping[vadKey]

			if !ok {
				vr = new(VadResult)
				vr.ScoreList = make([]*EmtionScore, 0)
				vadMapping[vadKey] = vr
				dcr.VadResults = append(dcr.VadResults, vr)
				vr.Status = int(status.Int64)
				vr.SegStartTime = st.Float64
				vr.SegEndTime = ed.Float64

				if blob != nil {
					var v map[string]interface{}
					json.Unmarshal(blob, &v)
					vr.ExtraInfo = v
				}
			}

			if emType.Valid && score.Valid {
				es := new(EmtionScore)
				es.Label = DefaultEmotion[int(emType.Int64)]
				es.Score = score.Float64
				vr.ScoreList = append(vr.ScoreList, es)
			}

		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return http.StatusInternalServerError, errors.New("Internal server error")
	}

	if count == 0 {
		return http.StatusBadRequest, errors.New("No file_id " + fileID)
	}

	var cr []*ChannelResult

	err = QueryChannelScore(id, &cr)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, errors.New("Internal server error")
	}

	for _, res := range cr {
		dcr, ok := channelMapping[res.ChannelID]
		if !ok {
			dcr = new(DetailChannelResult)
			channelMapping[res.ChannelID] = dcr
			drb.Channels = append(drb.Channels, dcr)
			dcr.ChannelID = res.ChannelID
		}
		dcr.Result = res.Result
	}

	return http.StatusOK, nil
}

/*
func QuerySingleResult(fileID string, appid string) ([]byte, int, error) {

	rb := new(ReturnBlock)
	rb.Channels = make([]*ChannelResult, 0)
	//query the fileInformation table
	query := SingleQuerySQL
	query += " where " + NFILEID + "=? and " + NAPPID + "=?"

	err := db.QueryRow(query, fileID, appid).Scan(&rb.FileID, &rb.FileName, &rb.FileType, &rb.Duration,
		&rb.CreateTime, &rb.Checksum, &rb.Tag, &rb.Priority, &rb.Size)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.StatusBadRequest, errors.New("No " + NFILEID + ": " + fileID)
		}
		log.Println(err)
		return nil, http.StatusInternalServerError, err
	}

	QueryChannelScore(fileID, &rb.Channels)

	//encode to json
	encodeRes, err := json.Marshal(rb)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	return encodeRes, http.StatusOK, nil

}
*/
func QueryResult(appid string, conditions string, conditions2 string, offset int, doPaging bool, rbs *[]*ReturnBlock) (int, int, error) {
	//debug.FreeOSMemory()

	query := "with joinr as ("
	query += QueryFileInfoAndChanScoreSQL + " from ( select * from " + MainTable
	query += " where " + NAPPID + "=?"
	query += " and " + conditions + " )"
	query += QueryFileInfoAndChanScoreSQL2
	if conditions2 != "" {
		query += " where " + conditions2
	}
	query += " )"
	query += "select * from joinr "
	if doPaging || offset > 0 {
		query += "where " + NFILEID + " in (select * from (select distinct " + NFILEID +
			" from joinr limit " + PAGELIMIT + " offset " + strconv.Itoa(offset) + ") as c);"
	}

	//log.Println(query)
	rows, err := db.Query(query, appid)
	if err != nil {
		log.Println(err)
		return 0, http.StatusInternalServerError, errors.New("Internal server error")
	}
	defer rows.Close()

	ReturnBlockMap := make(map[string]*ReturnBlock)
	ChannelMap := make(map[string]*ChannelResult)
	var count int
	for rows.Next() {
		nrb := new(ReturnBlock)
		nrb.Channels = make([]*ChannelResult, 0)
		var cr *ChannelResult
		var es *EmtionScore

		var channel sql.NullInt64
		var label sql.NullInt64
		var score sql.NullFloat64

		err := rows.Scan(&nrb.FileID, &nrb.FileName, &nrb.FileType, &nrb.Duration,
			&nrb.CreateTime, &nrb.Checksum, &nrb.Tag, &nrb.Tag2, &nrb.Priority, &nrb.Size, &nrb.AnalysisResult, &nrb.UploadTime,
			&channel, &label, &score)

		if err != nil {
			log.Println(err)
			return 0, http.StatusInternalServerError, errors.New("Internal server error")
		}

		// render write in ms convert to second.
		nrb.Duration /= 1000

		rb, ok := ReturnBlockMap[nrb.FileID]
		if !ok {
			rb = nrb
			*rbs = append(*rbs, rb)
			ReturnBlockMap[rb.FileID] = rb
			count++
		}

		if channel.Valid {
			key := rb.FileID + strconv.Itoa(int(channel.Int64))
			ch, ok := ChannelMap[key]
			if !ok {
				cr = new(ChannelResult)
				cr.ChannelID = int(channel.Int64)
				rb.Channels = append(rb.Channels, cr)
				ChannelMap[key] = cr
			} else {
				cr = ch
			}
			es = new(EmtionScore)
			es.Label = EmotionMap[int(label.Int64)]
			es.Score = score.Float64
			cr.Result = append(cr.Result, es)

		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return 0, http.StatusInternalServerError, errors.New("Internal server error")
	}
	return count, http.StatusOK, nil

}

func QueryChannelScore(id uint64, chs *[]*ChannelResult) error {
	//query the channelScore table
	query := QueryChannelScoreSQL
	query += " where " + NID + "=?" + " order by " + NCHANNEL

	rows, err := db.Query(query, id)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	lastCh := -1
	var cr *ChannelResult
	for rows.Next() {
		var channel int
		var emotionType int
		es := new(EmtionScore)
		err := rows.Scan(&channel, &emotionType, &es.Score)
		if err != nil {
			log.Println(err)
		} else {
			es.Label = EmotionMap[emotionType]
			if channel != lastCh {
				if lastCh != -1 {
					*chs = append(*chs, cr)
				}
				cr = new(ChannelResult)
				cr.ChannelID = channel
				lastCh = channel
			}
			cr.Result = append(cr.Result, es)
		}
	}
	if cr != nil {
		*chs = append(*chs, cr)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

func QueryReport(appid string, conditions string, rc *ReportCtx) error {

	if rc == nil {
		return nil
	}

	query := QueryReportSQL
	query += " where " + NAPPID + "=?" + " and " + conditions

	rows, err := db.Query(query, appid)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	rr := new(ReportRow)

	for rows.Next() {
		//rr.Count++
		err := rows.Scan(&rr.FileName, &rr.Duration, &rr.Tag1, &rr.Tag2, &rr.UploadT, &rr.ProcessST, &rr.ProcessET)
		if err != nil {
			log.Println(err)
			return err
		}

		err = rc.PutRecord(rr)
		if err != nil {
			log.Println(err)
			return err
		}

	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

func QueryCount(appid string, conditions string, conditions2 string) (int, error) {
	var count int

	query := "select count(distinct a.id) from ( select * from " + MainTable
	query += " where " + NAPPID + "=?"
	query += " and " + conditions + " )"
	query += QueryFileInfoAndChanScoreSQL2
	if conditions2 != "" {
		query += " where  " + conditions2
	}

	//log.Println(conditions2)
	//log.Println(query)

	rows, err := db.Query(query, appid)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Println(err)
			return 0, err
		}
	}
	return count, nil
}

func InitEmotionMap() {

	rows, err := db.Query(QueryEmotionMapSQL)
	if err != nil {
		EmotionMap = DefaultEmotion
		log.Println(err)
		log.Println("using default emotion map")
		return

	}
	defer rows.Close()

	EmotionMap = make(map[int]string)
	//var angerType string
	for rows.Next() {
		var id int
		var emotion string
		err := rows.Scan(&id, &emotion)
		if err != nil {
			EmotionMap = DefaultEmotion
			log.Println("using default emotion map")
			log.Println(err)
			return
		}

		EmotionMap[id] = emotion
		/*
			if emotion == "anger" {
				angerType = strconv.Itoa(id)
			}
		*/
	}

	if len(EmotionMap) == 0 {
		EmotionMap = DefaultEmotion
		log.Println("No emotion map in database.Using default emotion map")
	}
	/*
		if angerType != "" {
			AngerType = angerType
		}
	*/
}
