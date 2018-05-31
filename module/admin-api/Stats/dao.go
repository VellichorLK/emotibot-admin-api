package Stats

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

const (
	TAG_TYPE_TABLE_FORMAT = "%s_tag_type"
	TAG_TABLE_FORMAT      = "%s_tag"
	RECORD_TABLE_FORMAT   = "%s_record"
	RAW_RECORD_TABLE      = "chat_record"
	RECORD_INFO_TABLE     = "static_record_info"
)

func getAuditList(appid string, input *AuditInput) ([]*AuditLog, error) {
	// Audit log is not splited by appid for now
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return nil, errors.New("DB is not inited")
	}

	columns := []string{"user_id", "ip_source", "create_time", "module", "operation", "content", "result"}

	conditions := []string{}
	args := []interface{}{}

	if input.Filter != nil && input.Filter.Module != "-1" {
		conditions = append(conditions, "module = ?")
		args = append(args, input.Filter.Module)
	}
	if input.Filter != nil && input.Filter.Operation != "-1" {
		conditions = append(conditions, "operation = ?")
		args = append(args, input.Filter.Operation)
	}
	if input.Filter != nil && input.Filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, input.Filter.UserID)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, input.Start)
	args = append(args, input.End)

	queryStr := fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc", strings.Join(columns, ","), strings.Join(conditions, " and "))
	util.LogTrace.Printf("Query for audit: %s", queryStr)
	util.LogTrace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*AuditLog{}
	for rows.Next() {
		temp := AuditLog{}
		rows.Scan(&temp.UserID, &temp.UserIP, &temp.CreateTime, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
		ret = append(ret, &temp)
	}

	return ret, nil
}

func getAuditListData(appid string, input *AuditInput, page int, listPerPage int, export bool) ([]*AuditLog, int, error) {
	// Audit log is not splited by appid for now
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return nil, 0, errors.New("DB is not inited")
	}

	util.LogTrace.Printf("Search for audit: %#v", input.Filter)

	columns := []string{"id", "user_id", "ip_source", "UNIX_TIMESTAMP(create_time)", "module", "operation", "content", "result"}

	conditions := []string{}
	args := []interface{}{}

	if input.Filter != nil && input.Filter.Module != "-1" {
		conditions = append(conditions, "module = ?")
		args = append(args, input.Filter.Module)
	}
	if input.Filter != nil && input.Filter.Operation != "-1" {
		conditions = append(conditions, "operation = ?")
		args = append(args, input.Filter.Operation)
	}
	if input.Filter != nil && input.Filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, input.Filter.UserID)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, input.Start)
	args = append(args, input.End)

	shift := (page - 1) * listPerPage
	queryStr := ""
	if export == true {
		queryStr = fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc", strings.Join(columns, ","), strings.Join(conditions, " and "))
	} else {
		queryStr = fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc limit ? offset ?", strings.Join(columns, ","), strings.Join(conditions, " and "))
		args = append(args, listPerPage)
		args = append(args, shift)
	}

	util.LogTrace.Printf("Query for audit: %s", queryStr)
	util.LogTrace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ret := []*AuditLog{}
	for rows.Next() {
		temp := AuditLog{}
		var id int
		var timestamp int64
		rows.Scan(&id, &temp.UserID, &temp.UserIP, &timestamp, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
		temp.CreateTime = time.Unix(timestamp, 0)
		ret = append(ret, &temp)
	}

	cnt, err := getAuditListCnt(appid, input)
	if err != nil {
		return nil, 0, err
	}

	return ret, cnt, nil
}

func getAuditListCnt(appid string, input *AuditInput) (int, error) {
	// Audit log is not splited by appid for now
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return 0, errors.New("DB is not inited")
	}

	conditions := []string{}
	args := []interface{}{}

	if input.Filter != nil && input.Filter.Module != "-1" {
		conditions = append(conditions, "module = ?")
		args = append(args, input.Filter.Module)
	}
	if input.Filter != nil && input.Filter.Operation != "-1" {
		conditions = append(conditions, "operation = ?")
		args = append(args, input.Filter.Operation)
	}
	if input.Filter != nil && input.Filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, input.Filter.UserID)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, input.Start)
	args = append(args, input.End)

	queryStr := fmt.Sprintf("SELECT COUNT(*) FROM audit_record WHERE %s", strings.Join(conditions, " and "))
	util.LogTrace.Printf("Query for audit: %s", queryStr)
	util.LogTrace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var ret int
	if rows.Next() {
		err = rows.Scan(&ret)
		if err != nil {
			return 0, err
		}
	}

	return ret, nil
}

func initStatDB(url string, user string, pass string, db string) (*sql.DB, error) {
	return util.InitDB(url, user, pass, db)
}

func getUnresolveQuestionsStatistic(appid string, start int64, end int64) ([]*StatRow, error) {
	mySQL := getStatsDB()
	if mySQL == nil {
		return nil, errors.New("DB is not inited")
	}

	table := fmt.Sprintf(RECORD_TABLE_FORMAT, appid)
	queryPart := fmt.Sprintf("SELECT r.user_q, COUNT(*) as cnt, MAX(r.answer), MAX(r.score), r.std_q FROM %s AS r LEFT JOIN %s AS info USING(unique_id)", table, RECORD_INFO_TABLE)
	condition := "WHERE info.qa_solved = -1 and r.created_time between FROM_UNIXTIME(?) and FROM_UNIXTIME(?) GROUP BY r.user_q, r.std_q ORDER BY cnt DESC"

	queryStr := queryPart + " " + condition

	util.LogTrace.Printf("Query for stats unresolve question: %s, with [%d, %d]", queryStr, start, end)
	rows, err := mySQL.Query(queryStr, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*StatRow{}
	for rows.Next() {
		temp := StatRow{}
		rows.Scan(&temp.UserQuery, &temp.Count, &temp.Answer, &temp.Score, &temp.StandardQuestion)
		util.LogTrace.Printf("==== %#v", temp)
		ret = append(ret, &temp)
	}
	return ret, nil
}
func getDialogCnt(appid string, start int64, end int64, tagType string, tag string) (int, int, error) {
	statsDB := getStatsDB()
	if statsDB == nil {
		return 0, 0, errors.New("statsDB is not inited")
	}
	statsTable := fmt.Sprintf(RECORD_TABLE_FORMAT, appid)

	statsTableCntSql := "SELECT user_id" +
		" FROM %s" +
		" WHERE created_time BETWEEN FROM_UNIXTIME(?) AND FROM_UNIXTIME(?) AND %s = ? AND app_id = ?"
	statsTableCntSql = fmt.Sprintf(statsTableCntSql, statsTable, tagType)

	rawTableCntSql := "SELECT user_id" +
		" FROM " + RAW_RECORD_TABLE +
		" WHERE app_id = ? AND created_time BETWEEN FROM_UNIXTIME(?) AND FROM_UNIXTIME(?)"
	rawTableCntSql = fmt.Sprintf(rawTableCntSql)
	rawTableCntSql += " AND custom_info LIKE ?"

	likeInput := "%%\"" + tagType + "\":\"" + tag + "\"%%"
	querySql := fmt.Sprintf("SELECT COUNT(DISTINCT(user_id)), COUNT(1) FROM (%s UNION ALL %s) tmp", statsTableCntSql, rawTableCntSql)
	userCntRet := 0
	totalCntRet := 0

	ansRows, err := statsDB.Query(querySql, start, end, tag, appid, appid, start, end, likeInput)
	if err != nil {
		return 0, 0, err
	}
	defer ansRows.Close()

	if ansRows.Next() {
		ansRows.Scan(&userCntRet, &totalCntRet)
	}
	return userCntRet, totalCntRet, nil
}
func getDialogOneDayStatistic(appid string, start int64, end int64, tagType string) (string, []DialogStatsData, error) {
	emotibotDB := util.GetMainDB()
	if emotibotDB == nil {
		return "", nil, errors.New("emotibotDB is not inited")
	}

	tagTypeTable := fmt.Sprintf(TAG_TYPE_TABLE_FORMAT, appid)
	tagTable := fmt.Sprintf(TAG_TABLE_FORMAT, appid)

	var typeNameRet string
	dataRet := []DialogStatsData{}

	queryTag := "SELECT Tag_Name, Type_Name" +
		" FROM %s t1" +
		" INNER JOIN %s t2 ON t1.Tag_Type = t2.Type_id" +
		" WHERE t2.Type_Code = '%s'"
	queryTagSql := fmt.Sprintf(queryTag, tagTable, tagTypeTable, tagType)

	tagRows, err := emotibotDB.Query(queryTagSql)
	if err != nil {
		return "", nil, err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		data := DialogStatsData{}
		tagRows.Scan(&data.Tag, &typeNameRet)
		data.Tag = strings.Replace(data.Tag, "#", "", -1)

		data.UserCnt, data.TotalCnt, err = getDialogCnt(appid, start, end, tagType, data.Tag)
		if err != nil {
			return "", nil, err
		}
		dataRet = append(dataRet, data)
	}

	return typeNameRet, dataRet, nil
}

func getChatRecords(appID string, start, end time.Time, users ...string) ([]statsRow, error) {
	query := "SELECT user_id, user_q, answer, brand, created_time FROM `" + appID + "_record` WHERE (`created_time` BETWEEN ? AND ?) "
	var input = []interface{}{start, end}
	for i, uid := range users {
		if i == 0 {
			query += " AND user_id = ? "
		} else {
			query += " OR user_id = >"
		}
		input = append(input, uid)
	}
	query += " ORDER BY created_time DESC"
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, fmt.Errorf("can not get db")
	}
	rows, err := db.Query(query, input...)
	if err != nil {
		return nil, fmt.Errorf("query failed, %v", err)
	}
	defer rows.Close()
	var tRows = []statsRow{}
	for rows.Next() {
		var userID, userQuestion, answer, brand string
		var createdTime time.Time
		rows.Scan(&userID, &userQuestion, &answer, &brand, &createdTime)
		tRows = append(tRows, statsRow{
			"user_id":           userID,
			"name":              brand,
			"input":             userQuestion,
			"output":            answer,
			"conversation_time": createdTime,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql err: %v", err)
	}
	return tRows, nil
}

func getFAQStats(appID string, start, end time.Time, brandName string, eq ...whereEqual) ([]statsRow, error) {
	query := "SELECT category, std_question, SUM(count), `name` FROM " + FAQStatsTable.Name + " WHERE (cache_day BETWEEN ? AND ?) AND `name` = ?"
	var input = []interface{}{start, end, brandName}
	for _, e := range eq {
		query += " AND " + e.ColName + " = ? "
		input = append(input, e.value)
	}
	query += " GROUP BY category, std_question, name"

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, fmt.Errorf("can not get db")
	}
	rows, err := db.Query(query, input...)
	if err != nil {
		util.LogError.Printf("query: %s\n", query)
		return nil, fmt.Errorf("query failed, %v", err)
	}
	defer rows.Close()
	var sRows = []statsRow{}
	for rows.Next() {
		var category, stdQuestion, qBrandName string
		var total_count int64
		rows.Scan(&category, &stdQuestion, &total_count, &qBrandName)
		var categories = make([]string, 5, 5)
		var r = make(statsRow)
		copy(categories, strings.Split(category, "/"))

		r["question_name"] = stdQuestion
		r["brand"] = qBrandName
		r["total_count"] = total_count
		r["hit_count"] = total_count
		r["accuracy"] = 1
		for i, c := range categories {
			catKey := "categoryL" + strconv.Itoa(i)
			r[catKey] = c
		}
		sRows = append(sRows, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql failed, %v", err)
	}

	return sRows, nil
}

//getTagValue create a tag_name & tag_code mapping.
func getTagValue(appID string, typ int) (map[string]string, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("can not get db")
	}
	rows, err := db.Query("SELECT Tag_Code, Tag_Name FROM "+appID+"_tag WHERE Tag_Type = ?", typ)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()
	var tags = make(map[string]string)
	for rows.Next() {
		var name, value string
		rows.Scan(&name, &value)
		value = strings.Trim(value, "#")
		tags[name] = value
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql error: %v", err)
	}
	return tags, nil
}
