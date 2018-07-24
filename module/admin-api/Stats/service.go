package Stats

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetAuditList(appid string, input *AuditInput) (*AuditRet, int, error) {
	list, totalCnt, err := getAuditListData(appid, input, input.Page, input.ListPerPage, input.Export)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := &AuditRet{
		TotalCount: totalCnt,
		Data:       list,
	}

	return ret, ApiError.SUCCESS, nil
}

func GetQuestionStatisticResult(appid string, day int, qType string) (*StatRet, int, error) {
	now := time.Now().Local()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	start := end - int64(day*24*60*60)

	var data []*StatRow
	var err error
	if qType == "unsolved" {
		data, err = getUnresolveQuestionsStatistic(appid, start, end)
	} else {
		return nil, ApiError.REQUEST_ERROR, errors.New("Unsupport type")
	}

	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := StatRet{
		Data: data,
	}

	return &ret, ApiError.SUCCESS, nil
}

func GetDialogOneDayStatistic(appid string, startTime int64, endTime int64, tagType string) (*DialogStatsRet, int, error) {

	ret := DialogStatsRet{}
	typeName, datas, err := getDialogOneDayStatistic(appid, startTime, endTime, tagType)

	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	if len(typeName) <= 0 {
		return &ret, ApiError.REQUEST_ERROR, errors.New("not found tag_type")
	}
	header := DialogStatsHeader{}
	header.Id = "tag"
	header.Text = typeName
	ret.TableHeader = append(ret.TableHeader, header)

	header.Id = "userCnt"
	header.Text = "机器人接入客户量"
	ret.TableHeader = append(ret.TableHeader, header)

	header.Id = "totalCnt"
	header.Text = "机器人接入会话量"
	ret.TableHeader = append(ret.TableHeader, header)

	ret.Data = datas
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return &ret, ApiError.SUCCESS, nil
}

type StatResponse struct {
	Headers []Column   `json:"table_header"`
	Data    []statsRow `json:"data"`
}

//GetChatRecords will return records of users
func GetChatRecords(appID string, start, end time.Time, users ...string) (StatResponse, error) {
	data, err := getChatRecords(appID, start, end, users...)
	if err != nil {
		return StatResponse{}, err
	}
	return StatResponse{Headers: ChatRecordTable.Columns, Data: data}, nil
}

func GetFAQStats(appID string, start, end time.Time, brandName, keyword string) (StatResponse, error) {
	var keywords = []whereEqual{}
	if keyword != "" {
		keywords = append(keywords, whereEqual{"std_question", keyword})
	}
	data, err := getFAQStats(appID, start, end, brandName, keywords...)
	if err != nil {
		return StatResponse{}, err
	}
	return StatResponse{
		Headers: FAQStatsTable.Columns,
		Data:    data,
	}, nil
}

// SessionCSV fetch sessions data and formatted as csv content bytes
func SessionCSV(appID string, condition SessionCondition) ([]byte, error) {
	sessions, err := getSessions(appID, condition)
	if err != nil {
		return nil, fmt.Errorf("fetch sessions content error, %v", err)
	}
	var buffer bytes.Buffer
	//UTF-8 BOM handle
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buffer)
	writer.Write([]string{"会话id", "使用者ID", "状态", "开始时间", "结束时间", "会话长度", "会话资讯", "备忘"})
	for _, s := range sessions {
		start := time.Unix(s.StartTime, 0).In(asiaTaipei).Format(time.RFC3339)
		end := time.Unix(s.EndTime, 0).In(asiaTaipei).Format(time.RFC3339)
		duration := strconv.FormatInt(s.Duration, 10)
		status := strconv.FormatInt(s.Status, 10)
		info, _ := json.Marshal(s.Information)
		writer.Write([]string{s.ID, s.UserID, status, start, end, duration, string(info), s.Notes})

	}
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("io error, %v", err)
	}

	writer.Flush()
	return buffer.Bytes(), nil
}

//GetSessions fetch sessions based on condition given
func GetSessions(appID string, condition SessionCondition) (int, []Session, error) {
	sessions, err := getSessions(appID, condition)
	if err != nil {
		return 0, nil, fmt.Errorf("fetch sessions content error, %v", err)
	}
	count, err := getSessionCount(appID, condition)
	if err != nil {
		return 0, nil, fmt.Errorf("fetch session size error, %v", err)
	}

	return count, sessions, nil
}

//GetRecords retrive slice of record from records table
func GetRecords(appID, sessionID string) ([]record, error) {
	records, err := records(appID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get records failed, %v", err)
	}
	return records, nil
}

//GetDetailCSV create csv binary with detail about session row and it's chat record
func GetDetailCSV(appID, sessionID string) ([]byte, error) {

	sessions, err := getSessions(appID, SessionCondition{ID: sessionID, Limit: &PageLimit{PageSize: 1}})
	if err != nil {
		return nil, fmt.Errorf("get sessions failed, %v", err)
	}
	if len(sessions) <= 0 {
		return nil, fmt.Errorf("can not found any row with %s", sessionID)
	}

	var buffer bytes.Buffer
	//UTF-8 BOM handle
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buffer)
	writer.Write([]string{"会话id", "使用者ID", "状态", "开始时间", "结束时间", "会话长度", "会话资讯", "备忘"})
	s := sessions[0]
	start := time.Unix(s.StartTime, 0).In(asiaTaipei).Format(time.RFC3339)
	end := time.Unix(s.EndTime, 0).In(asiaTaipei).Format(time.RFC3339)
	duration := strconv.FormatInt(s.Duration, 10)
	status := strconv.FormatInt(s.Status, 10)
	info, _ := json.Marshal(s.Information)
	writer.Write([]string{s.ID, s.UserID, status, start, end, duration, string(info), s.Notes})

	writer.Write([]string{"对话时间", "客户问题", "机器人答复"})
	records, err := records(appID, sessionID)
	fmt.Printf("%+v", records)
	if err != nil {
		return nil, fmt.Errorf("get records failed, %v", err)
	}
	for _, r := range records {
		timeStr := time.Unix(r.Timestamp, 0).In(asiaTaipei).Format(time.RFC3339)
		writer.Write([]string{timeStr, r.UserText, r.RobotText})
	}
	if err = writer.Error(); err != nil {
		return nil, fmt.Errorf("csv io failed, %v", err)
	}
	fmt.Println(buffer.Bytes())
	writer.Flush()

	return buffer.Bytes(), nil
}
