package qi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	uuid "github.com/satori/go.uuid"
)

func CallsHandler(w http.ResponseWriter, r *http.Request) {
	query, err := newCallQuery(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("request error: %v", err))
		return
	}
	calls, err := Calls(*query)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get call failed, %v", err))
		return
	}
	util.WriteJSON(w, calls)
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

func NewCallsHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		CallID int64 `json:"call_id"`
	}
	req, err := extractNewCallReq(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("request error: %v", err))
		return
	}
	_ = req
	guid, err := uuid.NewV4()
	if err != nil {
		//Use 999 for now, we dont have general backend error.
		util.ReturnError(w, 999, fmt.Sprintf("generate uuid failed, %v", err))
		return
	}
	_ = guid
	// hyphenslessID := hex.EncodeToString(guid[:])
	// call := model.Call{
	// 	UUID:           hyphenslessID,
	// 	Description:    &req.CallComment,
	// 	UploadUnixTime: time.Now().Unix(),
	// 	CallUnixTime:   req.CallTime,
	// 	StaffID:        req.HostID,
	// 	StaffName:      req.HostName,
	// 	Ext:            req.Extension,
	// 	Department:     req.Department,
	// 	CustomerID:     req.GuestID,
	// 	CustomerName:   req.GuestName,
	// 	CustomerPhone:  req.GuestPhone,
	// 	EnterpriseID:   requestheader.GetEnterpriseID(r),
	// 	UploadUser:     requestheader.GetUserID(r),
	// 	Type:           model.CallTypeWholeFile,
	// 	LeftChanRole:   callRoleTyp(req.LeftChannel),
	// 	RightChanRole:  callRoleTyp(req.RightChannel),
	// }
	// id, err := NewCall(call)
	// if err != nil {
	// 	util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("creating call from failed, %v", err))
	// 	return
	// }
	// resp := response{CallID: id}
	// util.WriteJSON(w, resp)
	resp := response{CallID: 3}
	util.WriteJSON(w, resp)
}

func UpdateCallsFileHandler(w http.ResponseWriter, r *http.Request) {
	callID, found := mux.Vars(r)["call_id"]
	if !found || callID == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require call_id in path"))
		return
	}
	id, err := strconv.ParseInt(callID, 10, 64)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("invalid call_id '%s', need to be int.", callID))
	}
	_ = id
	// Calls(model.CallQuery{
	// 	ID: []int64{id},
	// })
	// err := r.ParseForm()
	// if err != nil {
	// 	util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("parse form failed, %v", err))
	// 	return
	// }
	// f, header, err := r.FormFile("upfile")
	// if err != nil {
	// 	util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("retrive upfile failed, %v", err))
	// 	return
	// }
	// defer f.Close()
	// //use 500 mb for limit now
	// if header.Size > (500 * 1024 * 1024) {
	// 	util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("upfile is over maximum size, %v", err))
	// 	return
	// }
	// if ext := path.Ext(header.Filename); ext != "wav" {
	// 	util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("extension '%s'is not valid, only wav is supported."), ext)
	// 	return
	// }
	// if volume == "" {
	// 	util.ReturnError(w, 999, fmt.Sprintf("volume is not exist, please contact ops and check init log for volume init error."))
	// }
	// filename := fmt.Sprintf("%s.wav", call_id)
	// volume.Open()
}

func CallsFileHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://ccrma.stanford.edu/~jos/mp3/gtr-nylon22.mp3")
	if err != nil {
		logger.Error.Println("Get err ", err)
		return
	}
	w.Header().Set("content-type", "audio/mpeg")
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error.Println(err)
		return
	}
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
	if _, found := callTypeDict[reqBody.LeftChannel]; !found {
		return nil, fmt.Errorf("request body's left channel value %s is not valid. the mapping should be %v", reqBody.LeftChannel, callTypeDict)
	}
	if _, found := callTypeDict[reqBody.RightChannel]; !found {
		return nil, fmt.Errorf("request body's right channel value %s is not valid. the mapping should be %v", reqBody.RightChannel, callTypeDict)
	}
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
}
