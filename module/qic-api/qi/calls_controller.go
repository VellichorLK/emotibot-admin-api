package qi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

// CallsDetailHandler fetch a single call with its segments.
func CallsDetailHandler(w http.ResponseWriter, r *http.Request, c *model.Call) {
	responses, _, err := CallRespsWithTotal(model.CallQuery{
		ID: []int64{c.ID},
	})
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("fetch task failed, %v", err))
	}
	resp := responses[0]
	resp.Segments, err = getSegments(*c)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get segment failed, %v", err))
		return
	}

	util.WriteJSON(w, resp)
}

func CallsHandler(w http.ResponseWriter, r *http.Request) {

	query, err := newModelCallQuery(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("request error: %v", err))
		return
	}
	calls, total, err := CallRespsWithTotal(*query)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get call failed, %v", err))
		return
	}
	resp := CallsResponse{
		Paging: general.Paging{
			Page:  query.Paging.Page,
			Limit: query.Paging.Limit,
			Total: total,
		},
		Data: calls,
	}
	util.WriteJSON(w, resp)
}

// NewCallsHandler create a call but no upload file itself.
func NewCallsHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		CallID int64 `json:"call_id"`
	}
	req, err := extractNewCallReq(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("request error: %v", err))
		return
	}

	call, err := NewCall(req)
	if err != nil || call == nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("creating call failed, %v", err))
		return
	}

	resp := response{CallID: call.ID}
	util.WriteJSON(w, resp)
}

func UpdateCallsFileHandler(w http.ResponseWriter, r *http.Request, c *model.Call) {
	f, header, err := r.FormFile("upfile")
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("retrive upfile failed, %v", err))
		return
	}
	defer f.Close()
	//use 500 mb for limit now
	if header.Size > (500 * 1024 * 1024) {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("upfile is over maximum size, %v", err))
		return
	}
	if ext := path.Ext(header.Filename); strings.ToLower(ext) != ".wav" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("extension '%s' is not valid, only wav is supported.", ext))
		return
	}
	if volume == "" {
		util.ReturnError(w, 999, fmt.Sprintf("volume is not exist, please contact ops and check init log for volume init error."))
		return
	}
	filename := fmt.Sprint(c.ID, ".wav")
	fp := fmt.Sprint(volume, "/", filename)

	outFile, err := os.Create(fp)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, fmt.Sprintf("create file failed, %v", err))
		return
	}
	_, err = io.Copy(outFile, f)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, fmt.Sprintf("write file failed, %v", err))
		return
	}
	// Volume only used in ourself, not expose to outside.
	c.FilePath = &filename
	err = ConfirmCall(c)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, fmt.Sprintf("confirm call failed, %v", err))
		return
	}
}

func CallsFileHandler(w http.ResponseWriter, r *http.Request, c *model.Call) {
	if c.FilePath == nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("file path has not set yet, pleasse check status before calling api"))
		return
	}
	fp := path.Join(volume, *c.FilePath)
	f, err := os.Open(fp)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, fmt.Sprintf("open file %s failed, %v", *c.FilePath, err))
	}

	w.Header().Set("content-type", "audio/mpeg")
	_, err = io.Copy(w, f)
	if err != nil {
		logger.Error.Println(err)
		util.ReturnError(w, AdminErrors.ErrnoIOError, fmt.Sprintf("open file %s failed, %v", *c.FilePath, err))
		return
	}
	return

}

// callRequest is a middleware for injecting call into next.
// it extract the call model from call_id path and enterprise header.
// any request error will terminate handler as BadRequest.
func callRequest(next func(w http.ResponseWriter, r *http.Request, c *model.Call)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const IDKey = "call_id"
		callID := mux.Vars(r)[IDKey]
		if callID == "" {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require %s in path", IDKey))
			return
		}
		id, err := strconv.ParseInt(callID, 10, 64)
		if err != nil {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("invalid call_id '%s', need to be int", callID))
			return
		}
		enterprise := requestheader.GetEnterpriseID(r)
		if enterprise == "" {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("empty enterprise ID"))
			return
		}
		c, err := Call(id, enterprise)
		if err == ErrNotFound {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("call_id '%s' is not exist", callID))
			return
		}
		if err != nil {
			util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("failed to query db, %v", err))
			return
		}
		next(w, r, &c)
	}
}

func extractNewCallReq(r *http.Request) (*NewCallReq, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body failed, %v", err)
	}
	defer r.Body.Close()
	reqBody := &NewCallReq{}
	err = json.Unmarshal(data, reqBody)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal failed, %v", err)
	}
	if reqBody.LeftChannel == "" {
		reqBody.LeftChannel = CallStaffRoleName
	}
	if reqBody.RightChannel == "" {
		reqBody.RightChannel = CallCustomerRoleName
	}
	if _, found := callTypeDict[reqBody.LeftChannel]; !found {
		return nil, fmt.Errorf("request body's left channel value %s is not valid. the mapping should be %v", reqBody.LeftChannel, callTypeDict)
	}
	if _, found := callTypeDict[reqBody.RightChannel]; !found {
		return nil, fmt.Errorf("request body's right channel value %s is not valid. the mapping should be %v", reqBody.RightChannel, callTypeDict)
	}
	enterprise := requestheader.GetEnterpriseID(r)
	if enterprise == "" {
		return nil, fmt.Errorf("enterpriseID is required")
	}
	reqBody.Enterprise = enterprise
	userID := requestheader.GetUserID(r)
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	reqBody.UploadUser = userID

	return reqBody, nil
}

// NewCallReq is the concrete request for creat a new call entity.
// 	all fields consider require columns for NewCallReq.
//	custom columns need to be checked valid for call's group.
type NewCallReq struct {
	FileName      string                 `json:"file_name"`
	CallTime      int64                  `json:"call_time"`
	CallComment   string                 `json:"call_comment"`
	Transaction   int64                  `json:"transaction"`
	Series        string                 `json:"series"`
	HostID        string                 `json:"host_id"`
	HostName      string                 `json:"host_name"`
	Extension     string                 `json:"extension"`
	Department    string                 `json:"department"`
	GuestID       string                 `json:"customer_id"`
	GuestName     string                 `json:"customer_name"`
	GuestPhone    string                 `json:"customer_phone"`
	LeftChannel   string                 `json:"left_channel"`
	RightChannel  string                 `json:"right_channel"`
	Enterprise    string                 `json:"-"`
	UploadUser    string                 `json:"-"`
	Type          int8                   `json:"-"`
	CustomColumns map[string]interface{} `json:"-"` //Custom columns of the call.
}

var callRequestJSONKeys = parseJSONKeys(NewCallReq{})

func parseJSONKeys(n interface{}) map[string]struct{} {
	ta := reflect.TypeOf(n)
	keys := map[string]struct{}{}
	for i := 0; i < ta.NumField(); i++ {
		tag := ta.Field(i).Tag.Get("json")
		key, _ := getJSONName(tag)
		if key == "-" {
			continue
		}
		keys[key] = struct{}{}
	}
	return keys
}

func getJSONName(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx:]
	} else {
		return tag, ""
	}
}

// UnmarshalJSON unmarshal request with optional custom columns
func (n *NewCallReq) UnmarshalJSON(data []byte) error {
	// Because we already overwrite the NewCalReq UnmarshalJSON,
	// use NewCallReq in json.Unmarshal here will cause looping.
	// the solution is to create an alias of the type.
	// check here: http://choly.ca/post/go-json-marshalling/
	// TODO: since we have all required key maybe we should use reflect to solve this issue here?
	type Alias NewCallReq
	a := &struct {
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	columns := map[string]interface{}{}
	if err := json.Unmarshal(data, &columns); err != nil {
		return err
	}

	for col, val := range columns {
		if _, exist := callRequestJSONKeys[col]; exist {
			continue
		}
		if n.CustomColumns == nil {
			n.CustomColumns = map[string]interface{}{}
		}
		n.CustomColumns[col] = val
	}
	return nil
}

type segment struct {
	ASRText    string       `json:"asr_text"`
	Emotion    []asrEmotion `json:"emotion"`
	Speaker    string       `json:"speaker"`
	StartTime  float64      `json:"start_time"`
	EndTime    float64      `json:"end_time"`
	SentenceID int64        `json:"sent_id"`
	SegmentID  int64        `json:"segment_id"`
	Status     int64        `json:"status"`
}

type asrEmotion struct {
	Type  string  `json:"type"`
	Score float64 `json:"score"`
}

type CallDetail struct {
	CallID   int64     `json:"call_id"`
	Status   int8      `json:"status"`
	Segment  []segment `json:"voice_result"`
	FileName string    `json:"file_name"`
}

type CallsResponse struct {
	Paging general.Paging `json:"paging"`
	Data   []CallResp     `json:"data"`
}

//CallResp is the UI struct of the call.
type CallResp struct {
	CallID           int64                  `json:"call_id"`
	CallTime         int64                  `json:"call_time"`
	Transaction      int8                   `json:"deal"`
	Status           int64                  `json:"status"`
	UploadTime       int64                  `json:"upload_time"`
	CallLength       float64                `json:"duration"`
	LeftSilenceTime  float64                `json:"left_silence_time"`
	RightSilenceTime float64                `json:"right_silence_time"`
	LeftSpeed        *float64               `json:"left_speed"`  // to compatible with old response
	RightSpeed       *float64               `json:"right_speed"` // to compatible with old response
	FileName         string                 `json:"file_name,omitempty"`
	CallComment      string                 `json:"call_comment,omitempty"`
	Series           string                 `json:"series,omitempty"`
	HostID           string                 `json:"staff_id,omitempty"`
	HostName         string                 `json:"staff_name,omitempty"`
	Extension        string                 `json:"extension,omitempty"`
	Department       string                 `json:"department,omitempty"`
	CustomerID       string                 `json:"customer_id,omitempty"`
	CustomerName     string                 `json:"customer_name,omitempty"`
	CustomerPhone    string                 `json:"customer_phone,omitempty"`
	LeftChannel      string                 `json:"left_channel,omitempty"`
	RightChannel     string                 `json:"right_channel,omitempty"`
	LeftSilenceRate  float64                `json:"left_silence_rate,omitempty"`
	RightSilenceRate float64                `json:"right_silence_rate,omitempty"`
	Segments         []segment              `json:"segments,omitempty"`
	CustomColumns    map[string]interface{} `json:"-"`
}

func (c CallResp) MarshalJSON() ([]byte, error) {
	resp := map[string]interface{}{}
	v := reflect.ValueOf(c)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name, opt := getJSONName(tag)
		if name == "-" {
			continue
		}
		if strings.Contains(opt, "omitempty") {
			f := v.Field(i)
			switch f.Kind() {
			case reflect.String:
				if f.String() == "" {
					continue
				}
			case reflect.Float64, reflect.Float32:
				if f.Float() == 0 {
					continue
				}
			case reflect.Int64, reflect.Int32, reflect.Int8:
				if f.Int() == 0 {
					continue
				}
			case reflect.Slice, reflect.Array, reflect.Map:
				if f.IsNil() {
					continue
				}
			}
		}

		resp[name] = v.Field(i).Interface()
	}
	for colName, val := range c.CustomColumns {
		if _, exist := resp[colName]; exist {
			return nil, fmt.Errorf("custom column %s is overlapped with require column", colName)
		}
		resp[colName] = val
	}
	return json.Marshal(resp)
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
