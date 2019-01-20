package qi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/qic-api/model/v1"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func CallsDetailHandler(w http.ResponseWriter, r *http.Request) {
	c := getCall(w, r)
	if c == nil {
		//already write error, just return
		return
	}
	results, err := getSegments(*c)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get segment failed, %v", err))
		return
	}
	//TODO: Detail need the result dao implemented
	resp := CallDetail{
		CallID:      c.ID,
		Status:      c.Status,
		VoiceResult: results,
		FileName:    *c.FileName,
		CuResult:    []interface{}{},
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

	id, err := NewCall(req)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("creating call failed, %v", err))
		return
	}
	resp := response{CallID: id}
	util.WriteJSON(w, resp)
}

func UpdateCallsFileHandler(w http.ResponseWriter, r *http.Request) {
	c := getCall(w, r)
	if c == nil {
		return
	}
	err := r.ParseForm()
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("parse request form failed, %v", err))
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
	c := getCall(w, r)
	if c == nil {
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
		return
	}
	return

}

func getCall(w http.ResponseWriter, r *http.Request) *model.Call {
	callID, found := mux.Vars(r)["call_id"]
	if !found || callID == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require call_id in path"))
		return nil
	}
	id, err := strconv.ParseInt(callID, 10, 64)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("invalid call_id '%s', need to be int.", callID))
		return nil
	}
	calls, err := Calls(nil, model.CallQuery{ID: []int64{id}})
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("failed to query call db, %v", err))
		return nil
	}
	if len(calls) == 0 {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("call_id '%s' is not exist", callID))
		return nil
	}
	c := calls[0]
	return &c
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

type NewCallReq struct {
	FileName     string `json:"file_name"`
	CallTime     int64  `json:"call_time"`
	CallComment  string `json:"call_comment"`
	Transaction  int64  `json:"transaction"`
	Series       string `json:"series"`
	HostID       string `json:"host_id"`
	HostName     string `json:"host_name"`
	Extension    string `json:"extension"`
	Department   string `json:"department"`
	GuestID      string `json:"guest_id"`
	GuestName    string `json:"guest_name"`
	GuestPhone   string `json:"guest_phone"`
	LeftChannel  string `json:"left_channel"`
	RightChannel string `json:"right_channel"`
	Enterprise   string `json:"-"`
	UploadUser   string `json:"-"`
}

type CallDetail struct {
	CallID      int64         `json:"call_id"`
	Status      int8          `json:"status"`
	VoiceResult []voiceResult `json:"voice_result"`
	FileName    string        `json:"file_name"`
	CuResult    []interface{} `json:"cu_result"`
}

// type cuResult struct {
// 	Score     float64    `json:"score"`
// 	QiResult  []QiResult `json:"qi_result"`
// 	GroupName string     `json:"group_name"`
// }

// type QiResult struct {
// 	ControllerRule string        `json:"controller_rule"`
// 	Valid          bool          `json:"valid"`
// 	Method         int64         `json:"method"`
// 	LogicResults   []LogicResult `json:"logic_results"`
// }

// type LogicResult struct {
// 	Label     []Label `json:"label"`
// 	LogicRule string  `json:"logic_rule"`
// 	Valid     bool    `json:"valid"`
// }

// type Label struct {
// 	Score *int64  `json:"score,omitempty"`
// 	ID    []int64 `json:"id"`
// 	Match []Match `json:"match"`
// 	Tag   string  `json:"tag"`
// 	Type  Type    `json:"type"`
// }

// type Match struct {
// 	Score    int64   `json:"score"`
// 	ID       int64   `json:"id"`
// 	Tag      string  `json:"tag"`
// 	Sentence string  `json:"sentence"`
// 	Match    *string `json:"match,omitempty"`
// }

type voiceResult struct {
	ASRText    string  `json:"asr_text"`
	Emotion    float64 `json:"emotion"`
	Speaker    string  `json:"speaker"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	SentenceID int64   `json:"sent_id"`
	Sret       int64   `json:"sret"` //status
}

// type Type string

// const (
// 	Intent  Type = "intent"
// 	Keyword Type = "keyword"
// )
