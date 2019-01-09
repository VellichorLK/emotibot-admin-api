package qi

import (
	"fmt"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	uuid "github.com/satori/go.uuid"
)

type CallResp struct {
	FileName         string  `json:"file_name,omitempty"`
	CallTime         int64   `json:"call_time,omitempty"`
	CallComment      string  `json:"call_comment,omitempty"`
	Transaction      int64   `json:"transaction,omitempty"`
	Series           string  `json:"series,omitempty"`
	HostID           string  `json:"host_id,omitempty"`
	HostName         string  `json:"host_name,omitempty"`
	Extension        string  `json:"extension,omitempty"`
	Department       string  `json:"department,omitempty"`
	GuestID          string  `json:"guest_id,omitempty"`
	GuestName        string  `json:"guest_name,omitempty"`
	GuestPhone       string  `json:"guest_phone,omitempty"`
	QiGroup          []int64 `json:"qi_group,omitempty"`
	LeftChannel      string  `json:"left_channel,omitempty"`
	RightChannel     string  `json:"right_channel,omitempty"`
	Status           int64   `json:"status,omitempty"`
	UploadTime       int64   `json:"upload_time,omitempty"`
	CallLength       float64 `json:"call_length,omitempty"`
	LeftSilenceRate  float64 `json:"left_silence_rate,omitempty"`
	RightSilenceRate float64 `json:"right_silence_rate,omitempty"`
	// ViolationNumber  int64   `json:"violation_number,omitempty"`
	// CriticalNumber   string  `json:"critical_number,omitempty"`
	// ReviewStatus     string  `json:"review_status,omitempty"`
	// Violation        string  `json:"violation,omitempty"`
	// CallScore        string  `json:"call_score,omitempty"`
	LeftSilenceTime  float64  `json:"left_silence_time"`
	RightSilenceTime float64  `json:"right_silence_time"`
	LeftSpeed        *float64 `json:"left_speed"`
	RightSpeed       *float64 `json:"right_speed"`
	// LeftAngry        float64 `json:"left_angry,omitempty"`
	// RightAngry       float64 `json:"right_angry,omitempty"`
}

type CallQueryRequest struct {
	Order       string
	Limit       int
	Page        int
	Content     *string
	StartTime   *int64
	EndTime     *int64
	Status      *int8
	Phone       *string
	Transcation *int
	Extention   *string
}

func Calls(request CallQueryRequest) ([]CallResp, error) {

	query := model.CallQuery{}
	if request.Status != nil {
		query.Status = []int8{*request.Status}
	}
	calls, err := callDao.Calls(nil, query)
	if err != nil {
		return nil, fmt.Errorf("call dao query failed, %v", err)
	}
	var result = make([]CallResp, 0, len(calls))
	for _, c := range calls {
		t, err := taskDao.CallTask(nil, c)
		if err != nil {
			return nil, fmt.Errorf("fetch task failed, %v", err)
		}
		var transaction int64 = 0
		if t.IsDeal {
			transaction = 1
		}
		r := CallResp{
			FileName:     *c.FileName,
			CallTime:     c.CallUnixTime,
			CallComment:  *c.Description,
			Transaction:  transaction,
			Series:       t.Series,
			HostID:       c.StaffID,
			HostName:     c.StaffName,
			Extension:    c.Ext,
			Department:   c.Department,
			GuestID:      c.CustomerID,
			GuestName:    c.CustomerName,
			GuestPhone:   c.CustomerPhone,
			LeftChannel:  callRoleTypStr(c.LeftChanRole),
			RightChannel: callRoleTypStr(c.RightChanRole),
			Status:       int64(c.Status),
			UploadTime:   c.UploadUnixTime,
			CallLength:   float64(c.DurationMillSecond) / 1000,
			LeftSpeed:    c.LeftSpeed,
			RightSpeed:   c.RightSpeed,
		}
		if c.LeftSilenceTime != nil {
			r.LeftSilenceTime = *c.LeftSilenceTime
			r.LeftSilenceRate = (r.LeftSilenceTime * 1000.0) / float64(c.DurationMillSecond)
		}
		if c.RightSilenceTime != nil {
			r.RightSilenceRate = *c.RightSilenceTime
			r.RightSilenceRate = (r.RightSilenceTime * 1000.0) / float64(c.DurationMillSecond)
		}
		result = append(result, r)
	}
	return result, nil
}

func NewCall(c model.Call) (int64, error) {
	_, err := uuid.FromString(c.UUID)
	if err != nil {
		return 0, fmt.Errorf("call UUID is not a valid uuid, %v", err)
	}
	calls, err := callDao.NewCalls(nil, []model.Call{c})
	if err == model.ErrAutoIDDisabled {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return calls[0].ID, nil
}

func newCallQuery(r *http.Request) (*CallQueryRequest, error) {
	var err error
	query := CallQueryRequest{}
	values := r.URL.Query()
	order := values.Get("order")
	if order == "" {
		return nil, fmt.Errorf("require order query string")
	}
	query.Order = order
	limit := values.Get("limit")
	if limit == "" {
		return nil, fmt.Errorf("require limit query string")
	}
	query.Limit, err = strconv.Atoi(limit)
	if err != nil {
		return nil, fmt.Errorf("limit is not a valid int, %v", err)
	}
	page := values.Get("page")
	if page == "" {
		return nil, fmt.Errorf("require page query string")
	}
	query.Page, err = strconv.Atoi(page)
	if err != nil {
		return nil, fmt.Errorf("page is not a valid int, %v", err)
	}
	if content := values.Get("content"); content != "" {
		query.Content = &content
	}
	if start := values.Get("start"); start != "" {
		startTime, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("start is not a valid int, %v", err)
		}
		query.StartTime = &startTime
	}
	if end := values.Get("end"); end != "" {
		endTime, err := strconv.ParseInt(end, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("end is not a valid int, %v", err)
		}
		query.EndTime = &endTime
	}
	if status := values.Get("status"); status != "" {
		statusTyp, err := strconv.ParseInt(status, 10, 8)
		statusInt8 := int8(statusTyp)
		if err != nil || callRoleTypStr(statusInt8) == "default" {
			return nil, fmt.Errorf("status is not a valid statu int.")
		}
		query.Status = &statusInt8
	}
	if phone := values.Get("phone"); phone != "" {
		query.Phone = &phone
	}
	if isTx := values.Get("transaction"); isTx != "" {
		transaction, err := strconv.Atoi(isTx)
		if err != nil || (transaction != 1 && transaction != 2) {
			return nil, fmt.Errorf("transaction is not a valid value")
		}
		query.Transcation = &transaction
	}
	if extension := values.Get("cs_phone"); extension != "" {
		query.Extention = &extension
	}
	return &query, nil
}

var callTypeDict = map[string]int8{
	"guest": model.CallChanCustomer,
	"host":  model.CallChanStaff,
}

func callRoleTyp(role string) int8 {
	value, found := callTypeDict[role]
	if !found {
		return model.CallChanDefault
	}
	return value
}
func callRoleTypStr(typ int8) string {
	for key, val := range callTypeDict {
		if val == typ {
			return key
		}
	}
	return "default"
}
