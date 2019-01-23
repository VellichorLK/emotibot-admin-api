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
		CallID:      c.ID,
		Status:      c.Status,
		VoiceResult: results,
		FileName:    *c.FileName,
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
}

type voiceResult struct {
	ASRText    string  `json:"asr_text"`
	Emotion    float64 `json:"emotion"`
	Speaker    string  `json:"speaker"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	SentenceID int64   `json:"sent_id"`
	Sret       int64   `json:"sret"` //status
}
