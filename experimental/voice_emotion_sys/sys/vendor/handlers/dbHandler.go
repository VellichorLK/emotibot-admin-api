package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

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

const QueryFileInfoAndChanScoreSQL = "select a." + NFILEID + ", a." + NFILENAME + ", a." + NFILETYPE + ", a." + NRDURATION +
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

	// 3. append the last one.
	mergedList = append(mergedList, VoiceSegment{
		STATUS_SILENCE,
		CHANNEL_SILENCE,
		voiceLength,
		voiceLength,
		nil,
		nil, 0, "", ""})

	// 4. use the merged list, the reverse part is the empty part.
	for index := len(mergedList) - 1; index >= 0; index-- {
		mergedList[index].Channel = CHANNEL_SILENCE
		mergedList[index].SegEndTime = mergedList[index].SegStartTime
		if index != 0 {
			mergedList[index].SegStartTime = mergedList[index-1].SegEndTime
		} else { // first item
			mergedList[index].SegStartTime = 0
		}
	}

	return mergedList
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

func ComputeChannelScore(eb *EmotionBlock) {

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

				if count > 10 && eb.RDuration > (120*1000) {
					upRate = float64(upRateCount) / float64(lastNSentence)

					primaryScore = 120*avgProb + 7 + 20*upRate

					if primaryScore >= 90 {

						s1 := rand.NewSource(time.Now().UnixNano())
						r1 := rand.New(s1)

						twoFixedScore = float64(90 + r1.Intn(6))
						if twoFixedScore < 95 {
							twoFixedScore += float64(r1.Intn(100)) / 100
						}
					} else {
						twoFixedScore = float64(int(primaryScore*100)) / 100
					}

				} else {
					twoFixedScore = float64(int((10+avgProb*10+7)*100)) / 100
				}

			}

			InsertChannelScore(eb.IDUint64, channel, emotionType, twoFixedScore)
		}
	}
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

	vads, err := getVadInfo(fileID)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, err
	}
	//debug.FreeOSMemory()
	drb.Channels = make([]*DetailChannelResult, 0)

	query := QueryDetailSQL

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

				if vadInfo, ok := vads[uint64(segID.Int64)]; ok {
					vr.Text = vadInfo.Text
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

	var cr []*ChannelResult

	err = QueryChannelScore(id, appid, &cr)
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

	cvs := make([]*ColumnValue, 0)
	err = QueryUsrFieldValue(id, appid, &cvs)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	drb.UsrColumn = cvs

	return http.StatusOK, nil
}

func getVadInfo(fileID string) (map[uint64]*vadInfo, error) {
	sqlString := "select " + NSEGID + "," + NVADRESULT + " from " + VadInfoTable +
		" where " + NVADFILEID + "=(select " + NID + " from " + MainTable + " where " + NFILEID + "=?)" +
		" and " + NPRIORITY + "=0 and " + NSTATUS + "=1"

	rows, err := db.Query(sqlString, fileID)
	if err != nil {
		return nil, err
	}

	fmt.Println(sqlString)
	fmt.Println(fileID)

	vads := make(map[uint64]*vadInfo)

	for rows.Next() {
		info := &vadInfo{}
		var segID uint64
		var result string
		err = rows.Scan(&segID, &result)
		if err != nil {
			return nil, err
		}
		info.Text = result
		vads[segID] = info
		//fmt.Printf("segID:%v, result:%s\n", segID, result)
	}

	return vads, nil
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
	var count int
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

		err := rows.Scan(&nrb.FileID, &nrb.FileName, &nrb.FileType, &nrb.Duration,
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
	return count, http.StatusOK, nil

}

func QueryChannelScore(id uint64, appid string, chs *[]*ChannelResult) error {
	//query the channelScore table
	query := QueryChannelScoreSQL
	query += " where " + NID + "= (select " + NID + " from " + MainTable + " where " + NFILEID + "=?)" + " order by " + NCHANNEL

	rows, err := db.Query(query, appid)
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
