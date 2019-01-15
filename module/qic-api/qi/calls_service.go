package qi

import (
	"fmt"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	uuid "github.com/satori/go.uuid"
)

//CallResp is the UI struct of the call.
type CallResp struct {
	CallID           int64   `json:"call_id"`
	FileName         string  `json:"file_name,omitempty"`
	CallTime         int64   `json:"call_time,omitempty"`
	CallComment      string  `json:"call_comment,omitempty"`
	Transaction      int64   `json:"deal,omitempty"`
	Series           string  `json:"series,omitempty"`
	HostID           string  `json:"staff_id,omitempty"`
	HostName         string  `json:"staff_name,omitempty"`
	Extension        string  `json:"extension,omitempty"`
	Department       string  `json:"department,omitempty"`
	GuestID          string  `json:"customer_id,omitempty"`
	GuestName        string  `json:"customer_name,omitempty"`
	GuestPhone       string  `json:"customer_phone,omitempty"`
	QiGroup          []int64 `json:"qi_group,omitempty"`
	LeftChannel      string  `json:"left_channel,omitempty"`
	RightChannel     string  `json:"right_channel,omitempty"`
	Status           int64   `json:"status,omitempty"`
	UploadTime       int64   `json:"upload_time,omitempty"`
	CallLength       float64 `json:"duration,omitempty"`
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

//CallQueryRequest is the input struct of func Calls
// type CallQueryRequest struct {
// 	ID          []int64
// 	Order       string
// 	Limit       int
// 	Page        int
// 	Content     *string
// 	StartTime   *int64
// 	EndTime     *int64
// 	Status      *int8
// 	Phone       *string
// 	Transcation *int
// 	Extention   *string
// }

func HasCall(id int64) (bool, error) {
	calls, err := callDao.Calls(nil, model.CallQuery{
		ID: []int64{id},
	})
	if err != nil {
		return false, fmt.Errorf("dao query failed, %v", err)
	}
	if len(calls) > 0 {
		return true, nil
	}
	return false, nil
}

func Calls(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error) {
	return callDao.Calls(delegatee, query)
}

//CallResps query the call and related information from different dao. and assemble it as a CallResp slice.
func CallResps(query model.CallQuery) (*CallsResponse, error) {
	if query.Paging == nil {
		return nil, fmt.Errorf("expect to have paging param")
	}
	total, err := callDao.Count(nil, query)
	if err != nil {
		return nil, fmt.Errorf("call dao count query failed, %v", err)
	}
	calls, err := Calls(nil, query)
	if err != nil {
		return nil, fmt.Errorf("call dao call query failed, %v", err)
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
			CallID:       c.ID,
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

	return &CallsResponse{
		Paging: general.Paging{
			Total: total,
			Limit: query.Paging.Limit,
			Page:  query.Paging.Page,
		},
		Data: result,
	}, nil
}

//NewCall create a call based on the input.
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

func UpdateCall(call *model.Call) error {
	return callDao.SetCall(nil, *call)
}

func newModelCallQuery(r *http.Request) (*model.CallQuery, error) {
	var err error
	query := model.CallQuery{}
	paging := &model.Pagination{}
	values := r.URL.Query()
	// order := values.Get("order")
	// if order == "" {
	// 	return nil, fmt.Errorf("require order query string")
	// }
	// paging.Order = order
	ent := requestheader.GetEnterpriseID(r)
	if ent == "" {
		return nil, fmt.Errorf("enterprise ID is required")
	}
	query.EnterpriseID = &ent
	limit := values.Get("limit")
	if limit == "" {
		return nil, fmt.Errorf("require limit query string")
	}
	paging.Limit, err = strconv.Atoi(limit)
	if err != nil {
		return nil, fmt.Errorf("limit is not a valid int, %v", err)
	}
	page := values.Get("page")
	if page == "" {
		return nil, fmt.Errorf("require page query string")
	}
	paging.Page, err = strconv.Atoi(page)
	if err != nil {
		return nil, fmt.Errorf("page is not a valid int, %v", err)
	}
	query.Paging = paging

	// if content := values.Get("content"); content != "" {
	// 	query.Content = &content
	// }
	if start := values.Get("start"); start != "" {
		startTime, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("start is not a valid int, %v", err)
		}
		query.CallTimeStart = &startTime
	}
	if end := values.Get("end"); end != "" {
		endTime, err := strconv.ParseInt(end, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("end is not a valid int, %v", err)
		}
		query.CallTimeEnd = &endTime
	}
	if statusGrp := values["status"]; len(statusGrp) > 0 {
		query.Status = make([]int8, 0, len(statusGrp))
		for _, status := range statusGrp {
			statusTyp, err := strconv.ParseInt(status, 10, 8)
			statusInt8 := int8(statusTyp)
			if err != nil || !model.ValidCallStatus(statusInt8) {
				return nil, fmt.Errorf("status %s is not a valid status flag", status)
			}
			query.Status = append(query.Status, statusInt8)
		}
	}
	if phone := values.Get("customer_phone"); phone != "" {
		query.CustomerPhone = &phone
	}
	if isTx := values.Get("deal"); isTx != "" {
		transaction, err := strconv.Atoi(isTx)
		if err != nil || (transaction != 0 && transaction != 1) {
			return nil, fmt.Errorf("deal is not a valid value")
		}
		txInt8 := int8(transaction)
		query.DealStatus = &txInt8
	}
	if extension := values.Get("extension"); extension != "" {
		query.Ext = &extension
	}
	if department := values.Get("department"); department != "" {
		query.Department = &department
	}
	return &query, nil
}

var callTypeDict = map[string]int8{
	CallStaffRoleName:    model.CallChanCustomer,
	CallCustomerRoleName: model.CallChanStaff,
}

const (
	CallStaffRoleName    = "staff"
	CallCustomerRoleName = "customer"
)

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
