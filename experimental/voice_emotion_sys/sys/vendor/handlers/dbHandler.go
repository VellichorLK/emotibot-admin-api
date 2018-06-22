package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

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

		log.Println(src)

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

const InsertUserFieldSQL = "insert into " + UsrColValTable + " (" + NID + "," + NCOLID + "," + NCOLVAL + ") values (?,?,?)"

const DeleteFileRowSQL = "delete from " + MainTable + " where " + NID + "=?"
const DeleteUsrFieldValueSQL = "delete from " + UsrColValTable + " where " + NID + "=?"

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

const QueryFileInfoAndChanScoreSQL = "select a." + NID + ",a." + NFILEID + ", a." + NFILENAME + ", a." + NFILETYPE + ", a." + NRDURATION +
	", a." + NFILET + ", a." + NCHECKSUM + ", a." + NTAG + ", a." + NTAG2 + ", a." + NPRIORITY + ", a." + NSIZE + ", a." + NANARES + ", a." + NUPT +
	", b." + NCHANNEL + ", b." + NEMOTYPE + ", b." + NSCORE + ", a." + NCOLID + ", a." + NCOLVAL

const QueryFileInfoAndChanScoreSQL2 = " as a left join " + ChannelTable + " as b on a." + NID + "=b." + NID

const QueryDetailSQL = "select * from (select " + NID + "," + NFILEID + "," + NFILENAME + "," + NFILETYPE + "," + NRDURATION + "," +
	NFILET + "," + NCHECKSUM + "," + NTAG + "," + NTAG2 + "," + NPRIORITY + "," + NSIZE + "," + NANARES + "," + NUPT +
	" from " + MainTable + " where " + NAPPID + "=?" + ") as a left join ( select b." + NSEGID + ",b." + NID + ",b." + NSEGST + ",b." + NSEGET +
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

	err = InsertDefaultUserFieldValue(fi.Appid, fi.ID)
	if err != nil {
		ExecSQL(DeleteFileRowSQL, lastID)
		return err
	}

	return nil
}

func InsertDefaultUserFieldValue(appid string, id string) error {
	dvs, ok := DefaulUsrField.DefaultValue[appid]
	if ok {
		stmt, err := db.Prepare(InsertUserFieldSQL)
		if err != nil {
			log.Println(err)
			return err
		}
		defer stmt.Close()

		for _, dv := range dvs {
			_, err := stmt.Exec(id, dv.ColID, dv.ColValue)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}

func ExecSQL(sql string, params ...interface{}) error {
	stmt, err := db.Prepare(sql)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(params...)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

/*
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

func DeleteUsrColumField(id string) error {
	sql := "delete from " + UsrColValTable + " where " + NID + "=?"
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

}
*/

func ComputeSilence(eb *EmotionBlock) error {

	if (nil == eb) || (nil == eb.Segments) || (0 == len(eb.Segments)) {
		e := errors.New("eb.segments is empty or nil. ")
		log.Println(e)
		return e
	}

	stmt, err := db.Prepare(InsertAnalysisSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	mergedList := GetSilenceList(eb)

	// Insert
	for _, aInterval := range mergedList {
		enc, err := json.Marshal(aInterval.ExtraInfo)
		if err != nil {
			log.Println(err)
			enc = []byte{}
		}

		_, err = stmt.Exec(eb.IDUint64, aInterval.SegStartTime, aInterval.SegEndTime, aInterval.Channel, aInterval.Status, enc)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func GetSilenceList(eb *EmotionBlock) []VoiceSegment {

	// 0. create silence list and do err check.
	silenceList := make([]VoiceSegment, 0)

	if eb == nil || eb.Segments == nil || len(eb.Segments) <= 0 {
		return silenceList
	}

	// 1. sort
	sort.Sort(ComparatorOfVoiceSegment(eb.Segments))

	// 2. merge
	var voiceLength float64 = float64(eb.RDuration) / 1000 //ms -> seconds
	mergedList := make([]VoiceSegment, 0, len(eb.Segments))
	for _, segmentInterval := range eb.Segments {
		if 0 == len(mergedList) || mergedList[len(mergedList)-1].SegEndTime < segmentInterval.SegStartTime {
			// non-overlap, must fit 0 <= startTime < endTime <= voiceLength
			if segmentInterval.SegStartTime <= segmentInterval.SegEndTime { // err check, most of cases start <= end
				mergedList = append(mergedList, VoiceSegment{
					segmentInterval.Status,
					segmentInterval.Channel,
					math.Max(0, segmentInterval.SegStartTime),
					math.Min(segmentInterval.SegEndTime, voiceLength),
					nil,
					nil, 0, segmentInterval.ID, ""})
			}
		} else {
			//overlap
			mergedList[len(mergedList)-1].SegEndTime = math.Max(mergedList[len(mergedList)-1].SegEndTime, segmentInterval.SegEndTime)
		}
	}

	const STATUS_SILENCE int = 1
	const CHANNEL_SILENCE int = 0

	// 3. caculate silence, overall merged = silence
	// 3.1 the voice is silence
	if len(mergedList) <= 0 {
		// after merged, the merged list is empty which means the overall voice is silence
		silenceList = append(silenceList, VoiceSegment{
			STATUS_SILENCE,
			CHANNEL_SILENCE,
			0,
			voiceLength,
			nil,
			nil, 0, "", ""})
		return silenceList
	}

	// 3.2 for each items in mergedList, caculate the silence part.
	for index := 0; index < len(mergedList); index++ {
		if (index == 0) && (mergedList[index].SegStartTime != 0) {
			// 3.2.1 first items in mergedList is not start from 0 then add silence part.
			silenceList = append(silenceList, VoiceSegment{
				STATUS_SILENCE,
				CHANNEL_SILENCE,
				0,
				mergedList[index].SegStartTime,
				nil,
				nil, 0, "", ""})
		} else if (0 < index) && (index < len(mergedList)) {
			// 3.2.3 others, Not first and last item
			silenceList = append(silenceList, VoiceSegment{
				STATUS_SILENCE,
				CHANNEL_SILENCE,
				mergedList[index-1].SegEndTime,
				mergedList[index].SegStartTime,
				nil,
				nil, 0, "", ""})
		}
	} //end for loop
	// 3.3 add the last silence
	if mergedList[len(mergedList)-1].SegEndTime != voiceLength {
		silenceList = append(silenceList, VoiceSegment{
			STATUS_SILENCE,
			CHANNEL_SILENCE,
			mergedList[len(mergedList)-1].SegEndTime,
			voiceLength,
			nil,
			nil, 0, "", ""})
	}

	return silenceList
}

// sort interface.
type ComparatorOfVoiceSegment []VoiceSegment

func (voiceSegmentlist ComparatorOfVoiceSegment) Len() int {
	return len(voiceSegmentlist)
}
func (voiceSegmentlist ComparatorOfVoiceSegment) Swap(i, j int) {
	voiceSegmentlist[i], voiceSegmentlist[j] = voiceSegmentlist[j], voiceSegmentlist[i]
}
func (voiceSegmentlist ComparatorOfVoiceSegment) Less(i, j int) bool {
	return voiceSegmentlist[i].SegStartTime < voiceSegmentlist[j].SegStartTime
}

func InsertAnalysisRecord(eb *EmotionBlock) error {
	stmt, err := db.Prepare(InsertAnalysisSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	for i, d := range eb.Segments {

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

		eb.Segments[i].SegmentID = lastID

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

func ComputeChannelScore(eb *EmotionBlock) []*TotalEmotionScore {

	const divider = 1000

	EmotionScoreInEachChannel := make(map[int][]float64)
	for _, seg := range eb.Segments {
		for _, s := range seg.ScoreList {
			label, ok := s.Label.(float64)
			if !ok {
				log.Printf("label is not float, %v %T\n", s.Label, s.Label)
			} else {
				emotionChannelKey := seg.Channel*divider + int(label)
				EmotionScoreInEachChannel[emotionChannelKey] = append(EmotionScoreInEachChannel[emotionChannelKey], s.Score)
			}
		}
	}

	emotionScore := make([]*TotalEmotionScore, 0, 4)

	for chEmotion, scores := range EmotionScoreInEachChannel {
		channel := chEmotion / divider
		emotionType := chEmotion % divider
		count := len(scores)

		if count > 0 {

			var twoFixedScore float64
			var weight float64
			var calculateNum float64
			//customer
			if channel == 1 {
				sort.Sort(sort.Reverse(sort.Float64Slice(scores)))
				if count > 30 && eb.RDuration > (180*1000) {

					const gap = 0.6

					const extractNum = 5

					hasEmotionCount := 0
					weight = 0.8
					for _, s := range scores {
						if s < gap {
							break
						}
						hasEmotionCount++
					}

					if hasEmotionCount > extractNum {
						for i := 0; i < extractNum; i++ {
							twoFixedScore += scores[i]
						}
						calculateNum = extractNum

						if twoFixedScore/calculateNum > 0.75 {
							weight = 0.64
						}

					} else {
						for _, s := range scores {
							twoFixedScore += s
						}
						calculateNum = float64(count)
					}

				} else {
					weight = 0.5
					for _, s := range scores {
						twoFixedScore += s
					}
					calculateNum = float64(count)
				}

				twoFixedScore = float64(int((twoFixedScore/calculateNum)*100*weight*100)) / 100
			} else {
				var upRateCount int
				var avgCountAcc float64
				var upRate, avgProb float64
				var primaryScore float64

				const gap = 0.6
				const weight = 1.5
				const portion = 0.25

				var lastNSentence int
				lastNSentence = int(math.Ceil(float64(count) / float64(4)))

				for i := 0; i < count; i++ {
					avgCountAcc += scores[i]
				}

				for i := 0; i < lastNSentence; i++ {
					if scores[count-1-i] > gap {
						upRateCount++
					}
				}

				avgProb = avgCountAcc / float64(count)

				upRate = float64(upRateCount) / float64(lastNSentence)

				primaryScore = 120*avgProb + 7 + 20*upRate

				if primaryScore >= 90 {
					primaryScore = 30*avgProb + 60 + 10*upRate
				}

				twoFixedScore = float64(int(primaryScore*100)) / 100

			}

			tes := &TotalEmotionScore{Channel: channel, EType: emotionType, Score: twoFixedScore}
			emotionScore = append(emotionScore, tes)

			InsertChannelScore(eb.IDUint64, channel, emotionType, twoFixedScore)
		}
	}

	return emotionScore
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

func QuerySingleDetail(fileID string, appid string, drb *DetailReturnBlock) (int, error) {

	enableASR := os.Getenv("ASR_ENABLE")
	enableASR = strings.ToLower(enableASR)

	var vads map[uint64]*VadInfo
	var err error

	if enableASR == "true" {
		//get vad information, prohibited words, text
		vads, err = getVadInfo(fileID, appid)
		if err != nil {
			log.Println(err)
			return http.StatusInternalServerError, err
		}
	}

	drb.Channels = make([]*DetailChannelResult, 0)

	query := QueryDetailSQL

	//get every segment of file
	rows, err := db.Query(query, appid, appid, fileID, fileID)
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
		var file, segID sql.NullInt64
		var blob []uint8

		err := rows.Scan(&id, &drb.FileID, &drb.FileName, &drb.FileType, &drb.Duration,
			&drb.CreateTime, &drb.Checksum, &drb.Tag, &drb.Tag2, &drb.Priority, &drb.Size, &drb.AnalysisResult, &drb.UploadTime,
			&segID, &file, &st, &ed, &ch, &status, &blob, &emType, &score)

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
				//because jason use channel 0 as silence period
				if chInt > 0 {
					drb.Channels = append(drb.Channels, dcr)
				}
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

				if vadInfo, ok := vads[uint64(segID.Int64)]; ok {
					vr.Text = vadInfo.Text
					vr.Prohibit = vadInfo.Prohibit
					vr.Speed = GetFloatPrecesion(vadInfo.Speed, 2)
				}

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

	//add silence period, jason use channel 0 as silence period
	if dcr, ok := channelMapping[0]; ok {
		num := len(dcr.VadResults)
		if num > 0 {
			drb.SilencePeriod = make([]*SilencePeriod, num)
		}
		for i := 0; i < num; i++ {
			sp := &SilencePeriod{}
			sp.SegStartTime = GetFloatPrecesion(dcr.VadResults[i].SegStartTime, 2)
			sp.SegEndTime = GetFloatPrecesion(dcr.VadResults[i].SegEndTime, 2)
			sp.Period = GetFloatPrecesion(sp.SegEndTime-sp.SegStartTime, 2)
			drb.SilencePeriod[i] = sp
		}
	}
	//query channel summary score
	var cr []*ChannelResult

	err = QueryChannelScore(fileID, appid, &cr)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, errors.New("Internal server error")
	}

	for _, res := range cr {
		dcr, ok := channelMapping[res.ChannelID]
		if ok {
			dcr.ChannelResult.Result = res.Result
		}
	}

	//add user defined colum value
	cvs := make([]*ColumnValue, 0)
	err = QueryUsrFieldValue(id, appid, &cvs)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	drb.UsrColumn = cvs

	return http.StatusOK, nil
}

func getVadInfo(fileID string, appid string) (map[uint64]*VadInfo, error) {

	wordsMap, err := getProhitbitWords(appid)
	if err != nil {
		return nil, err
	}

	sqlString := "select " + NSEGID + "," + NVADRESULT + "," + NPROHIBITID + "," + NSPEECHSPEED + " from " + VadInfoTable +
		" where " + NVADFILEID + "=(select " + NID + " from " + MainTable + " where " + NFILEID + "=? and " + NAPPID + "=?" + ")" +
		" and " + NPRIORITY + "=0 and " + NSTATUS + "=1"

	rows, err := db.Query(sqlString, fileID, appid)
	if err != nil {
		return nil, err
	}

	vads := make(map[uint64]*VadInfo)

	for rows.Next() {

		var segID, prohibitID uint64
		var result string
		var speechSpeed float64
		err = rows.Scan(&segID, &result, &prohibitID, &speechSpeed)
		if err != nil {
			return nil, err
		}
		info := &VadInfo{Text: result, Speed: speechSpeed}
		vads[segID] = info
		if prohibitID > 0 {
			if words, ok := wordsMap[prohibitID]; ok {
				info.Prohibit = words
			} else {
				log.Printf("[Warning] fileID:%s appid:%s segmentID:%v has prohibited ID:%v, but cannot find its words\n", fileID, appid, segID, prohibitID)
			}
		}

	}

	return vads, nil
}

func getProhitbitWords(appid string) (map[uint64]string, error) {
	querySQL := fmt.Sprintf("select %s,%s from %s where %s=?", NID, NPROHIBIT, ProhibitedTable, NJAPPID)
	rows, err := db.Query(querySQL, appid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	prohibitedMap := make(map[uint64]string)
	var id uint64
	var words string
	for rows.Next() {
		rows.Scan(&id, &words)
		prohibitedMap[id] = words
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return prohibitedMap, nil
}

func QueryResult(appid string, conditions string, conditions2 string, offset int, doPaging bool, checkLimit *CheckLimit, rbs *[]*ReturnBlock) (int, int, error) {
	//debug.FreeOSMemory()

	tmpRes := "(" + QueryFileInfoAndChanScoreSQL + " from ( select a.*,b." + NCOLID + ",b." + NCOLVAL + " from " + MainTable
	tmpRes += " as a left join " + UsrColValTable + " as b on a." + NID + "=b." + NID
	tmpRes += " where " + NAPPID + "=?"
	tmpRes += " and " + conditions + " )"
	tmpRes += QueryFileInfoAndChanScoreSQL2
	if conditions2 != "" {
		tmpRes += " where " + conditions2
	}
	tmpRes += ")"

	//log.Println(tmpRes)
	params := make([]interface{}, 0)
	params = append(params, appid)

	query := "select * from " + tmpRes + " as y " //left join " + UsrColValTable + "as z on y." + NID + "=z." + NID
	if doPaging || offset > 0 {
		query += "where " + NFILEID + " in (select " + NFILEID + " from (select x." + NFILEID +
			" from " + tmpRes + " as x  group by " + NFILEID + " order by " + NFILET + " desc" +
			" limit " + PAGELIMIT + " offset " + strconv.Itoa(offset) + ") as c)"
		params = append(params, appid)
	}
	query += " order by " + NFILET + " desc"

	//log.Println(query)
	rows, err := db.Query(query, params...)
	if err != nil {
		log.Println(err)
		return 0, http.StatusInternalServerError, errors.New("Internal server error")
	}
	defer rows.Close()

	ReturnBlockMap := make(map[string]*ReturnBlock)
	ChannelMap := make(map[string]*ChannelResult)
	EmotionUsedMap := make(map[string]bool)
	ColumFieldMap := make(map[string]bool)
	idList := make([]interface{}, 0)
	idToFileID := make(map[uint64]string)
	var count int
	True := true
	False := false
	var id uint64
	for rows.Next() {
		nrb := new(ReturnBlock)
		nrb.Channels = make([]*ChannelResult, 0)
		nrb.UsrColumn = make([]*ColumnValue, 0)
		var cr *ChannelResult
		var es *EmtionScore

		var channel sql.NullInt64
		var label sql.NullInt64
		var score sql.NullFloat64
		var colID sql.NullString
		var colValue sql.NullString

		err := rows.Scan(&id, &nrb.FileID, &nrb.FileName, &nrb.FileType, &nrb.Duration,
			&nrb.CreateTime, &nrb.Checksum, &nrb.Tag, &nrb.Tag2, &nrb.Priority, &nrb.Size, &nrb.AnalysisResult, &nrb.UploadTime,
			&channel, &label, &score, &colID, &colValue)

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
			idList = append(idList, id)
			idToFileID[id] = nrb.FileID
			rb.OverSilence = &False
		}

		if channel.Valid {
			key := rb.FileID + strconv.Itoa(int(channel.Int64))
			ch, ok := ChannelMap[key]
			if !ok {
				cr = new(ChannelResult)
				cr.ChannelID = int(channel.Int64)
				rb.Channels = append(rb.Channels, cr)
				ChannelMap[key] = cr
				cr.OverProhitbited = &False
				cr.OverSpeed = &False
			} else {
				cr = ch
			}

			emoKey := key + strconv.Itoa(int(label.Int64))
			_, ok = EmotionUsedMap[emoKey]
			if !ok {
				es = new(EmtionScore)
				es.Label = EmotionMap[int(label.Int64)]
				es.Score = score.Float64
				cr.Result = append(cr.Result, es)
				EmotionUsedMap[emoKey] = true
			}

		}

		if colID.Valid && colValue.Valid {
			key := rb.FileID + colID.String
			_, ok := ColumFieldMap[key]
			if !ok {
				ColumFieldMap[key] = true

				name, ok := DefaulUsrField.FieldNameMap[colID.String]
				if ok {
					val := &ColumnValue{name, colValue.String, colID.String}
					rb.UsrColumn = append(rb.UsrColumn, val)
				}

			}
		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return 0, http.StatusInternalServerError, errors.New("Internal server error")
	}

	enableASR := os.Getenv("ASR_ENABLE")
	enableASR = strings.ToLower(enableASR)

	//check the over list
	if len(idList) > 0 && enableASR == "true" {

		//check silence over the limit
		silenceLimit := checkLimit.SilenceLimitDuration
		overSilenceIDs, err := getOverSilence(silenceLimit, idList)
		if err != nil {
			log.Println(err)
			return 0, http.StatusInternalServerError, errors.New("Internal server error")
		}
		for id, count := range overSilenceIDs {
			if count > checkLimit.SilenceLimitCount {
				rb := ReturnBlockMap[idToFileID[id]]
				rb.OverSilence = &True
			}
		}

		//check speed, prohitbied words

		//get total count for each channel
		segCount, err := getSegCount(idList)
		if err != nil {
			log.Println(err)
			return 0, http.StatusInternalServerError, errors.New("Internal server error")
		}

		//get over speed
		speedLimit := checkLimit.SpeedLimit
		overSpeedSeg, err := getOverSpeed(speedLimit, idList)
		if err != nil {
			log.Println(err)
			return 0, http.StatusInternalServerError, errors.New("Internal server error")
		}
		for id, chMap := range overSpeedSeg {
			for ch, count := range chMap {
				if float64(count)/float64(segCount[id][ch]) > float64(checkLimit.SpeedLimitPercent)/100 {
					key := idToFileID[id] + strconv.Itoa(ch)
					chInfo := ChannelMap[key]
					chInfo.OverSpeed = &True
				}
			}
		}

		//get over prohibited words
		prohibitedLimit := checkLimit.ProhibitedLimit
		overProhibitedSeg, err := getOverProhibitedWords(idList)
		if err != nil {
			log.Println(err)
			return 0, http.StatusInternalServerError, errors.New("Internal server error")
		}

		for id, chMap := range overProhibitedSeg {
			for ch, count := range chMap {
				if count > prohibitedLimit {
					key := idToFileID[id] + strconv.Itoa(ch)
					chInfo := ChannelMap[key]
					chInfo.OverProhitbited = &True
				}
			}
		}

	}

	return count, http.StatusOK, nil
}

func getOverSilence(limitSilence int, ids []interface{}) (map[uint64]int, error) {
	silenceSQL := fmt.Sprintf("select %s,count(%s) from %s where %s=0 and %s-%s>=? and %s in (?%s) group by %s",
		NID, NID, AnalysisTable, NCHANNEL, NSEGET, NSEGST, NID, strings.Repeat(",?", len(ids)-1), NID)

	params := make([]interface{}, 0)
	params = append(params, limitSilence)
	params = append(params, ids...)

	rows, err := db.Query(silenceSQL, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	overIDs := make(map[uint64]int)
	var id uint64
	var count int
	for rows.Next() {
		err = rows.Scan(&id, &count)
		if err != nil {
			return nil, err
		}
		overIDs[id] = count
	}

	return overIDs, nil

}

func getOverSpeed(limitSpeed int, ids []interface{}) (map[uint64]map[int]int, error) {

	speedSQL := fmt.Sprintf("select b.%s, a.%s, count(b.%s) from %s  as a inner join %s as b on a.%s=b.%s where b.%s>? and b.%s in (?%s) group by a.%s,b.%s",
		NVADFILEID, NCHANNEL, NVADFILEID, AnalysisTable, VadInfoTable, NSEGID, NSEGID, NSPEECHSPEED, NVADFILEID, strings.Repeat(",?", len(ids)-1), NCHANNEL, NVADFILEID)

	params := make([]interface{}, 0)
	params = append(params, limitSpeed)
	params = append(params, ids...)

	return getSegMap(speedSQL, params)
}

func getOverProhibitedWords(ids []interface{}) (map[uint64]map[int]int, error) {
	speedSQL := fmt.Sprintf("select b.%s, a.%s, count(b.%s) from %s  as a inner join %s as b on a.%s=b.%s where b.%s>0 and b.%s in (?%s) group by a.%s,b.%s",
		NVADFILEID, NCHANNEL, NVADFILEID, AnalysisTable, VadInfoTable, NSEGID, NSEGID, NPROHIBITID, NVADFILEID, strings.Repeat(",?", len(ids)-1), NCHANNEL, NVADFILEID)

	return getSegMap(speedSQL, ids)
}

func getSegCount(ids []interface{}) (map[uint64]map[int]int, error) {
	segCountSQL := fmt.Sprintf("select b.%s, a.%s, count(b.%s) from %s  as a inner join %s as b on a.%s=b.%s where b.%s in (?%s) group by a.%s,b.%s",
		NVADFILEID, NCHANNEL, NVADFILEID, AnalysisTable, VadInfoTable, NSEGID, NSEGID, NVADFILEID, strings.Repeat(",?", len(ids)-1), NCHANNEL, NVADFILEID)
	return getSegMap(segCountSQL, ids)
}

func getSegMap(querySQL string, params []interface{}) (map[uint64]map[int]int, error) {
	segChMap := make(map[uint64]map[int]int)
	rows, err := db.Query(querySQL, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id uint64
	var ch, total int
	for rows.Next() {
		err = rows.Scan(&id, &ch, &total)
		if err != nil {
			return nil, err
		}
		segCount, ok := segChMap[id]
		if !ok {
			segCount = make(map[int]int)
			segChMap[id] = segCount
		}
		segCount[ch] = total
	}

	return segChMap, nil
}

func QueryChannelScore(id string, appid string, chs *[]*ChannelResult) error {
	//query the channelScore table
	query := QueryChannelScoreSQL
	query += " where " + NID + "= (select " + NID + " from " + MainTable + " where " + NFILEID + "=? and " + NAPPID + "=?)" + " order by " + NCHANNEL

	rows, err := db.Query(query, id, appid)
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

func QueryUsrFieldValue(id uint64, appid string, cvs *[]*ColumnValue) error {

	nameMap, err := getUsrColumNameMap(appid)
	if err != nil {
		return err
	}

	sql := "select " + NCOLID + "," + NCOLVAL + " from " + UsrColValTable + " where " + NID + "=?"
	rows, err := db.Query(sql, id)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var colID, colVal string
		err := rows.Scan(&colID, &colVal)
		if err != nil {
			log.Println(err)
			return err
		}

		colIDuint64, err := strconv.ParseUint(colID, 10, 64)
		if err != nil {
			return err
		}
		name, ok := nameMap[colIDuint64]
		if !ok {
			return errors.New("No col_id " + colID + " name")
		}

		cv := &ColumnValue{name, colVal, colID}
		*cvs = append(*cvs, cv)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

}

func getUsrColumNameMap(appid string) (map[uint64]string, error) {
	sqlString := "select " + NCOLID + "," + NCOLNAME + " from " + UsrColTable + " where " + NAPPID + "=?"
	rows, err := db.Query(sqlString, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nameMap := make(map[uint64]string)
	var name string
	var colID uint64
	for rows.Next() {
		err := rows.Scan(&colID, &name)
		if err != nil {
			return nil, err
		}
		nameMap[colID] = name
	}
	return nameMap, nil

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

//GetAppidByID get appid from id
func GetAppidByID(id uint64) (string, error) {
	querySQL := fmt.Sprintf("select %s from %s where %s=?", NAPPID, MainTable, NID)

	rows, err := db.Query(querySQL, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var appid string

	if rows.Next() {
		err = rows.Scan(&appid)
	}
	return appid, err
}

//GetColumnByID get value by assigned column
func GetColumnByID(columns []string, id uint64) ([]interface{}, error) {

	count := len(columns)

	if count == 0 {
		return nil, nil
	}

	querySQL := "select "

	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	for idx, v := range columns {
		if idx != 0 {
			querySQL += ","
		}
		querySQL += v
		valuePtrs[idx] = &values[idx]
	}

	querySQL += " from " + MainTable + " where " + NID + "=?"

	rows, err := db.Query(querySQL, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(valuePtrs...)
	}
	return values, err

}

//QuerySQL
func QuerySQL(querySQL string, params ...interface{}) (*sql.Rows, error) {
	return db.Query(querySQL, params...)
}
