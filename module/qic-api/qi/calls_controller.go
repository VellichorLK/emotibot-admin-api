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

func CallsDetailHandler(w http.ResponseWriter, r *http.Request) {
	c, err := callRequest(r)
	if ae, ok := err.(adminError); ok {
		util.ReturnError(w, ae.ErrorNo(), fmt.Sprintf("call request error: %v", ae))
		return
	} else if err != nil {
		//Unknown error
		util.ReturnError(w, 666, fmt.Sprintf("unknown error, %v", err))
		return
	}
	results, err := getSegments(*c)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get segment failed, %v", err))
		return
	}
	//TODO: Detail need the result dao implemented
	resp := CallDetail{
		CallID:   c.ID,
		Status:   c.Status,
		Segment:  results,
		FileName: *c.FileName,
	}
	util.WriteJSON(w, resp)
}

func CallsHandler(w http.ResponseWriter, r *http.Request) {

	query, err := newModelCallQuery(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("request error: %v", err))
		return
	}
	resp, err := CallResps(*query)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get call failed, %v", err))
		return
	}

	util.WriteJSON(w, resp)
}

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

func UpdateCallsFileHandler(w http.ResponseWriter, r *http.Request) {
	c, err := callRequest(r)
	if ae, ok := err.(adminError); ok {
		util.ReturnError(w, ae.ErrorNo(), ae.Error())
		return
	} else if err != nil {
		util.ReturnError(w, 666, fmt.Sprintf("unknown error: %v", err))
		return
	}

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

func CallsFileHandler(w http.ResponseWriter, r *http.Request) {
	c, err := callRequest(r)
	if ae, ok := err.(adminError); ok {
		util.ReturnError(w, ae.ErrorNo(), ae.Error())
		return
	} else if err != nil {
		util.ReturnError(w, 666, fmt.Sprintf("unknown error: %v", err))
		return
	}
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

// callRequest receive the request and get the call model from call_id.
// an error will returned if failure happened.
// the error return should be compatible to adminError interface.
func callRequest(r *http.Request) (*model.Call, error) {
	const IDKey = "call_id"
	callID := mux.Vars(r)[IDKey]
	if callID == "" {
		return nil, controllerError{
			error: fmt.Errorf("require %s in path", IDKey),
			errNo: AdminErrors.ErrnoRequestError,
		}
	}
	id, err := strconv.ParseInt(callID, 10, 64)
	if err != nil {
		return nil, controllerError{
			error: fmt.Errorf("invalid call_id '%s', need to be int", callID),
			errNo: AdminErrors.ErrnoRequestError,
		}
	}
	enterprise := requestheader.GetEnterpriseID(r)
	if enterprise == "" {
		return nil, controllerError{
			error: fmt.Errorf("empty enterprise ID"),
			errNo: AdminErrors.ErrnoRequestError,
		}
	}
	calls, err := Calls(nil, model.CallQuery{ID: []int64{id}, EnterpriseID: &enterprise})
	if err != nil {
		return nil, controllerError{
			error: fmt.Errorf("failed to query call db, %v", err),
			errNo: AdminErrors.ErrnoDBError,
		}
	}
	if len(calls) < 1 {
		return nil, controllerError{
			error: fmt.Errorf("call_id '%s' is not exist", callID),
		}
	}
	return &calls[0], nil
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
		var key string
		f := ta.Field(i)
		tag := f.Tag.Get("json")
		if idx := strings.Index(tag, ","); idx != -1 {
			key = tag[:idx]
		} else {
			key = tag
			if key == "-" {
				continue
			}
		}
		keys[key] = struct{}{}
	}
	return keys
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
	CallID           int64     `json:"call_id"`
	CallTime         int64     `json:"call_time"`
	Transaction      int8      `json:"deal"`
	Status           int64     `json:"status"`
	UploadTime       int64     `json:"upload_time"`
	CallLength       float64   `json:"duration"`
	LeftSilenceTime  float64   `json:"left_silence_time"`
	RightSilenceTime float64   `json:"right_silence_time"`
	LeftSpeed        *float64  `json:"left_speed"`
	RightSpeed       *float64  `json:"right_speed"`
	FileName         string    `json:"file_name,omitempty"`
	CallComment      string    `json:"call_comment,omitempty"`
	Series           string    `json:"series,omitempty"`
	HostID           string    `json:"staff_id,omitempty"`
	HostName         string    `json:"staff_name,omitempty"`
	Extension        string    `json:"extension,omitempty"`
	Department       string    `json:"department,omitempty"`
	CustomerID       string    `json:"customer_id,omitempty"`
	CustomerName     string    `json:"customer_name,omitempty"`
	CustomerPhone    string    `json:"customer_phone,omitempty"`
	LeftChannel      string    `json:"left_channel,omitempty"`
	RightChannel     string    `json:"right_channel,omitempty"`
	LeftSilenceRate  float64   `json:"left_silence_rate,omitempty"`
	RightSilenceRate float64   `json:"right_silence_rate,omitempty"`
	Segments         []segment `json:"segments"`
}
