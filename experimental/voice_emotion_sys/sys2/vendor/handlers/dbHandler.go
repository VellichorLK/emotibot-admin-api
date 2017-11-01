package handlers

import (
	"database/sql"
	"encoding/json"
	"log"

	"errors"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

//InitDatabaseCon create the initial connection to database
func InitDatabaseCon(ip string, port string, username string, password string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if db == nil {
		src := username + ":" + password + "@tcp(" + ip + ":" + port + ")/" + DataBase
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
	NDURATION + ", " + NFILET + ", " + NCHECKSUM + ", " + NPRIORITY + ", " + NAPPID + ", " + NSIZE + ", " +
	NFILEPATH + ", " + NUPT + ", " + NRDURATION + ")" + " values(?,?,?,?,?,?,?,?,?,?,?,?)"

const InsertUserFieldSQL = "insert into " + UsrColValTable + " (" + NID + "," + NCOLID + "," + NCOLVAL + ") values (?,?,?)"

const DeleteFileRowSQL = "delete from " + MainTable

//InsertAnalysisSQL used for insert a new emotion record
const InsertAnalysisSQL = "insert into " + AnalysisTable + " (" + NID + ", " + NSEGST + ", " + NSEGET + ", " +
	NCHANNEL + ", " + NSTATUS + ", " + NEXTAINFO + " )" + " values(?,?,?,?,?,?)"

//InsertEmotionSQL used for insert a new emotion record
const InsertEmotionSQL = "insert into " + EmotionTable + " (" + NSEGID + ", " + NEMOTYPE + ", " + NSCORE + " )" +
	" values(?,?,?)"

//QueryEmotionMapSQL query the emotion map table, only do once at the statup
const QueryEmotionMapSQL = "select " + NEMOTYPE + ", " + NEMOTION + " from " + EmotionMapTable

//Render sql
const (
	InsertChannelScoreSQL = "insert into " + ChannelTable + " (" + NID + ", " + NCHANNEL + ", " + NEMOTYPE + ", " + NSCORE + ") " +
		" values(?,?,?,?)"
	UpdateResultSQL = "update " + MainTable + " set " + NANAST + "=? ," + NANAET + "=? ," + NANARES + "=? ," + NRDURATION +
		"=? where " + NID + "=?"
)

//Query sql for api /files
const (
	QueryFileInfo = "select a." + NID + ", a." + NFILEID + ", a." + NFILENAME + ", a." + NFILETYPE + ", a." + NRDURATION +
		", a." + NFILET + ", a." + NCHECKSUM + ", a." + NPRIORITY + ", a." + NAPPID + ", a." + NSIZE + ", a." + NANARES + ",a." + NUPT +
		", b." + NCHANNEL + ", b." + NEMOTYPE + ", b." + NSCORE + ", c." + NTAGID + ", c." + NTAG + ", d." + NCOLID + ", d." + NCOLVAL + " from " + MainTable + " as a left join " +
		ChannelTable + " as b on a." + NID + "=b." + NID + " left join " + UserDefinedTagsTable + " as c on a." + NID + "=c." + NID +
		" left join " + UsrColValTable + " as d on a." + NID + "=d." + NID

	QueryDetailSQL = "select * from (select " + NID + "," + NFILEID + "," + NFILENAME + "," + NFILETYPE + "," + NRDURATION + "," +
		NFILET + "," + NCHECKSUM + "," + NPRIORITY + "," + NSIZE + "," + NANARES + "," + NUPT +
		" from " + MainTable + ") as a left join ( select b." + NID + ",b." + NSEGST + ",b." + NSEGET +
		",b." + NCHANNEL + ",b." + NSTATUS + ",b." + NEXTAINFO + ",c." + NEMOTYPE + ",c." + NSCORE + " from ( select * from " +
		AnalysisTable + " where " + NID + "=(select " + NID + " from " + MainTable + " where " +
		NAPPID + "=? and " + NFILEID + "=?)) as b left join " + EmotionTable + " as c " +
		"on b." + NSEGID + "=c." + NSEGID + ") as d on a." + NID + "=d." + NID + " where a." + NFILEID + "=?"

	QueryChannelScoreSQL = "select " + NCHANNEL + ", " + NEMOTYPE + ", " + NSCORE + " from " + ChannelTable
	QueryAnalysisSQL     = "select " + NEXTAINFO + " from " + AnalysisTable
)

const QueryReportSQL = "select " + NFILENAME + ", " + NRDURATION + ", " + NFILET + ", " + NANAST + "," + NANAET + " from " + MainTable

//Tag operation SQL
const (
	DeleteTagsSQL       = "delete from " + UserDefinedTagsTable
	InsertUserTag       = "insert into " + UserDefinedTagsTable + " (" + NID + ", " + NTAG + ") values(?,?)"
	QueryTagsByAppidSQL = "select distinct " + NTAG + " from " + UserDefinedTagsTable + " where " + NID +
		" in ( select " + NID + " from " + MainTable + " where " + NAPPID + " = ?)"
	QueryTagsByIDSQL     = "select " + NTAG + " from " + UserDefinedTagsTable + " where " + NID + "=?"
	DeleteTagByFileIDSQL = "delete from " + UserDefinedTagsTable + " where " + NID +
		" = ( select " + NID + " from " + MainTable + " where " + NAPPID + "=? and " + NFILEID + " = ?) and " + NTAG + " = ?"
	AddTagByFileIDSQL = "insert into " + UserDefinedTagsTable + "( " + NID + "," + NTAG + ") values(?,?)"
	GetTagByFileIDSQL = "select " + NID + ", " + NTAG + " from " + UserDefinedTagsTable + " where " + NID +
		" = ( select " + NID + " from " + MainTable + " where " + NAPPID + "=? and " + NFILEID + " = ?)"
	UpdateTagByFileIDSQL = "update " + UserDefinedTagsTable + " set " + NTAG + " = ? where " + NID +
		" = ( select " + NID + " from " + MainTable + " where " + NAPPID + "=? and " + NFILEID + " = ?)" +
		" and " + NTAG + "=?"
	UpdateTagByAppidSQL = "update " + UserDefinedTagsTable + " set " + NTAG + " = ? where " + NID +
		" in ( select " + NID + " from " + MainTable + " where " + NAPPID + "=?)" +
		" and " + NTAG + "=?"
	GetIDByFileIDSQL = "select " + NID + " from " + MainTable + " where " + NAPPID + " =? and " + NFILEID + " =?"
)

func InsertFileRecord(fi *FileInfo, tx *sql.Tx) error {

	res, err := ExecuteSQL(tx, InsertFileInfoSQL, fi.FileID, fi.FileName, fi.FileType, fi.Duration, fi.CreateTime,
		fi.Checksum, fi.Priority, fi.Appid, fi.Size, fi.Path, fi.UPTime, 0)

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

	if len(fi.UsrColumn) > 0 {
		err = InsertUsrFieldValue(fi, tx)
	}

	return nil
}

func InsertUsrFieldValue(fi *FileInfo, tx *sql.Tx) error {
	stmt, err := tx.Prepare(InsertUserFieldSQL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()
	for _, uc := range fi.UsrColumn {
		_, err := stmt.Exec(fi.ID, uc.ColID, uc.Value)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil

}

func InsertUserDefinedTags(id string, tags []string, tx *sql.Tx) error {
	var stmt *sql.Stmt
	var err error
	if tx == nil {
		stmt, err = db.Prepare(InsertUserTag)
	} else {
		stmt, err = tx.Prepare(InsertUserTag)
	}

	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	for _, tag := range tags {
		_, err := stmt.Exec(id, tag)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func UpdateTags(query string, params ...interface{}) (sql.Result, error) {
	return ExecuteSQL(nil, query, params...)
}

func GetIDByFileID(appid string, fileID string) (string, error) {
	rows, err := db.Query(GetIDByFileIDSQL, appid, fileID)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer rows.Close()

	var id string
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
			return "", err
		}
	}

	return id, nil
}

//GetTagsByFileID return id, tags, error
func GetTagsByFileID(appid string, fileID string) (string, []string, error) {
	rows, err := db.Query(GetTagByFileIDSQL, appid, fileID)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}
	defer rows.Close()

	tags := make([]string, 0, LIMITUSERTAGS)

	var id string
	for rows.Next() {
		var tag string
		err := rows.Scan(&id, &tag)
		if err != nil {
			log.Println(err)
			return "", nil, err
		}
		tags = append(tags, tag)
	}

	//when no tag for this file
	if id == "" {
		id, err = GetIDByFileID(appid, fileID)
		if err != nil {
			return "", nil, err
		}
	}

	return id, tags, nil

}

func ExecuteSQL(tx *sql.Tx, query string, params ...interface{}) (sql.Result, error) {

	var stmt *sql.Stmt
	var err error

	if tx == nil {
		stmt, err = db.Prepare(query)
	} else {
		stmt, err = tx.Prepare(query)
	}

	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(params...)
}

func DeleteFileRecord(id uint64) (sql.Result, error) {
	query := DeleteFileRowSQL
	query += " where " + NID + "=?"
	return ExecuteSQL(nil, query, id)
}

func DeleteTag(id uint64) (sql.Result, error) {
	query := DeleteTagsSQL
	query += " where " + NID + "=?"
	return ExecuteSQL(nil, query, id)
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
			}

		}
	}

	for chEmotion, count := range NumSegInOneChannel {
		var twoFixedScore, totalScore, weightScore, rateScore float64
		weightScore = TotalProbabilityWithEmotion[chEmotion] / float64(count)
		rateScore = float64(NumOfHasOneEmotionInOneChannel[chEmotion]) / float64(count)
		totalScore = weightScore*weight + rateScore*(1-weight)
		//log.Printf("id:%v, ch:%d, emotion:%d, score:%v, chemotion:%v\n", eb.IDUint64, chEmotion/divider, chEmotion%divider, totalScore*100, chEmotion)

		twoFixedScore = float64(int(totalScore*100*100)) / 100

		InsertChannelScore(eb.IDUint64, chEmotion/divider, chEmotion%divider, twoFixedScore)
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

func QuerySingleDetail(fileID string, appid string, drb *DetailReturnBlock) (int, error) {

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
			&drb.CreateTime, &drb.Checksum, &drb.Priority, &drb.Size, &drb.AnalysisResult, &drb.UploadTime,
			&file, &st, &ed, &ch, &status, &blob, &emType, &score)

		if err != nil {
			log.Println(err)
			return http.StatusInternalServerError, errors.New("Internal server error")
		}

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

	//make real_duration unit as second
	drb.Duration = drb.Duration / 1000

	if count == 0 {
		return http.StatusBadRequest, errors.New("No file_id " + fileID)
	}

	//query channel score table
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

	//query tags

	tags, err := QueryTags(QueryTagsByIDSQL, id)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	drb.Tags = tags

	cvs := make([]*ColumnValue, 0)
	err = QueryUsrFieldValue(id, &cvs)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	drb.UsrColumn = cvs

	return http.StatusOK, nil
}

func QueryResult(offset int64, conditions string, params ...interface{}) ([]*ReturnBlock, int, error) {

	tmpRes := "(" + QueryFileInfo + " where " + NAPPID + "=?"
	if conditions != "" {
		tmpRes += conditions
	}
	tmpRes += ")"

	query := "select * from " + tmpRes + "as y where " + NFILEID +
		" in (select " + NFILEID + " from (select x." + NFILEID +
		" from " + tmpRes + "as x group by " + NFILEID + " order by " + NFILET + " desc" +
		" limit " + PAGELIMIT + " offset " + strconv.FormatInt(offset, 10) + ") as d)" +
		" order by " + NFILET + " desc"

	doubleParams := make([]interface{}, 0)
	doubleParams = append(doubleParams, params...)
	doubleParams = append(doubleParams, params...)
	//log.Println(query)
	//log.Println(doubleParams...)
	rows, err := db.Query(query, doubleParams...)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	defer rows.Close()

	rbs := make([]*ReturnBlock, 0)

	ReturnBlockMap := make(map[string]*ReturnBlock)
	ChannelMap := make(map[string]*ChannelResult)
	LabelMap := make(map[string]bool)
	TagsMap := make(map[string]bool)
	UsrColMap := make(map[string]bool)

	for rows.Next() {
		nrb := new(ReturnBlock)
		nrb.Channels = make([]*ChannelResult, 0)
		nrb.Tags = make([]string, 0)
		var cr *ChannelResult
		var es *EmtionScore

		var channel sql.NullInt64
		var label sql.NullInt64
		var score sql.NullFloat64
		var tagID sql.NullString
		var tag sql.NullString
		var appid string
		var id int64

		var colID sql.NullString
		var colValue sql.NullString

		err := rows.Scan(&id, &nrb.FileID, &nrb.FileName, &nrb.FileType, &nrb.Duration,
			&nrb.CreateTime, &nrb.Checksum, &nrb.Priority, &appid, &nrb.Size, &nrb.AnalysisResult, &nrb.UploadTime,
			&channel, &label, &score, &tagID, &tag, &colID, &colValue)
		if err != nil {
			log.Println(err)
			return nil, http.StatusInternalServerError, errors.New("Internal server error")
		}

		rb, ok := ReturnBlockMap[nrb.FileID]
		if !ok {
			rb = nrb

			//make real_duration unit as second
			nrb.Duration = nrb.Duration / 1000

			rbs = append(rbs, rb)
			ReturnBlockMap[rb.FileID] = rb
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

			labelKey := key + strconv.FormatInt(label.Int64, 10)
			_, ok = LabelMap[labelKey]
			if !ok {
				es = new(EmtionScore)
				es.Label = EmotionMap[int(label.Int64)]
				es.Score = score.Float64
				cr.Result = append(cr.Result, es)
				LabelMap[labelKey] = true
			}

		}

		if tag.Valid {
			_, ok := TagsMap[tagID.String]
			if !ok {
				rb.Tags = append(rb.Tags, tag.String)
				TagsMap[tagID.String] = true
			}

		}

		if colID.Valid {
			key := rb.FileID + colID.String
			_, ok := UsrColMap[key]
			if !ok {
				UsrColMap[key] = true
				//nameInterface, ok := DefaulUsrField.FieldNameMap.Load(colID.String)

				//if ok {
				//name := nameInterface.(string)
				val := &ColumnValue{Value: colValue.String, ColID: colID.String}
				rb.UsrColumn = append(rb.UsrColumn, val)
				//}
			}
		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}

	return rbs, http.StatusOK, nil
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

func QueryUsrFieldValue(id uint64, cvs *[]*ColumnValue) error {
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
		/*
			nameInterface, ok := DefaulUsrField.FieldNameMap.Load(colID)

			if !ok {
				return errors.New("No col_id " + colID + " name")
			}

			name := nameInterface.(string)
		*/
		cv := &ColumnValue{Value: colVal, ColID: colID}
		*cvs = append(*cvs, cv)
	}

	err = rows.Err()
	if err != nil {
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

	//log.Println(query)

	rows, err := db.Query(query, appid)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	rr := new(ReportRow)

	for rows.Next() {
		//rr.Count++
		err := rows.Scan(&rr.FileName, &rr.Duration, &rr.UploadT, &rr.ProcessST, &rr.ProcessET)
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

func QueryCount(conditions string, params ...interface{}) (int64, error) {
	var count int64

	query := "select count(distinct " + NFILEID + ") from ("
	query += QueryFileInfo + " where " + NAPPID + "=?"

	if conditions != "" {
		query += conditions
	}
	query += ") as d"

	rows, err := db.Query(query, params...)
	if err != nil {
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
		//log.Println(err)
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
	}

	if len(EmotionMap) == 0 {
		EmotionMap = DefaultEmotion
		log.Println("No emotion map in database.Using default emotion map")
	}

}

func QueryTags(query string, identifier interface{}) ([]string, error) {
	tags := make([]string, 0, LIMITUSERTAGS)

	rows, err := db.Query(query, identifier)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			log.Println(err)
		} else {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}
