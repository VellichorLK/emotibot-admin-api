package Stats

import (
	"errors"
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

//GetSessions fetch sessions based on condition given
func GetSessions(appID string, condition SessionCondition) (int, []Session, error) {
	return getSessions(appID, condition)
}
