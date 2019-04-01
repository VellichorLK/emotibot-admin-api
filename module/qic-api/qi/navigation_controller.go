package qi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/sensitive"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/module/qic-api/util/timecache"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

type reqCallInIntent struct {
	IntentName string   `json:"intent_name"`
	Role       string   `json:"role"`
	Type       string   `json:"type"`
	Sentences  []string `json:"sentences"`
}

type respDetailFlow struct {
	Name       string                    `json:"name"`
	IntentName string                    `json:"intent_name"`
	Role       string                    `json:"role"`
	Type       string                    `json:"type"`
	Sentences  []model.SimpleSentence    `json:"sentences"`
	Nodes      []SentenceGroupInResponse `json:"nodes"`
}

var callInIntentMap = map[string]int{
	"":       0,
	"fixed":  0,
	"intent": 1,
}

var callInIntentCodeMap = map[int]string{
	0: "fixed",
	1: "intent",
}

func detailFlowToSetting(d *DetailNavFlow) *respDetailFlow {
	if d == nil {
		return &respDetailFlow{Sentences: []model.SimpleSentence{}, Nodes: []SentenceGroupInResponse{}}
	}
	resp := &respDetailFlow{Name: d.NavFlow.Name}

	if d.SentenceGroup.Sentences == nil {
		resp.Sentences = []model.SimpleSentence{}
	} else {
		resp.Sentences = d.SentenceGroup.Sentences
	}

	if d.IgnoreIntent == 1 {
		resp.Type = callInIntentCodeMap[0]
	} else {
		resp.Type = callInIntentCodeMap[1]
	}
	resp.IntentName = d.IntentName
	resp.Role = roleCodeMap[d.SentenceGroup.Role]

	if d.Nodes == nil {
		resp.Nodes = []SentenceGroupInResponse{}
	} else {
		resp.Nodes = make([]SentenceGroupInResponse, 0, len(d.Nodes))
		for _, v := range d.Nodes {
			r := sentenceGroupToSentenceGroupInResponse(&v)
			resp.Nodes = append(resp.Nodes, r)
		}
	}

	return resp
}

func handleGetFlowSetting(w http.ResponseWriter, r *http.Request) {

	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	enterprise := requestheader.GetEnterpriseID(r)
	setting, err := GetFlowSetting(id, enterprise)
	if err != nil {
		logger.Error.Printf("get the flow setting failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if setting == nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}
	resp := detailFlowToSetting(setting)
	util.WriteJSON(w, resp)
}

type reqNewFlow struct {
	Name       string `json:"name"`
	IntentName string `json:"intent_name"`
	Type       string `json:"type"`
}

func handleNewFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	var req reqNewFlow
	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	id, err := NewFlow(&req, enterprise)
	if err != nil {
		logger.Error.Printf("create the flow failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		ID int64 `json:"nav_id,string"`
	}{ID: id})
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}

//RespFlowList is the respose structure for flow list
type RespFlowList struct {
	Page pageResp       `json:"paging"`
	Data []*FlowSummary `json:"data"`
}

//FlowSummary is the response data unit for flow list
type FlowSummary struct {
	ID         int64  `json:"nav_id,string"`
	Name       string `json:"name"`
	IntentName string `json:"intent_name"`
	NumOfNodes int64  `json:"num_nodes"`
}

var navTypeMap = map[string]int{
	"fixed": 1,
}

func handleFlowList(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	typ := r.URL.Query().Get("type")

	isDelete := 0
	q := &model.NavQuery{Enterprise: &enterprise, IsDelete: &isDelete}
	if typCode, ok := navTypeMap[typ]; ok {
		q.IgnoreIntent = &typCode
	} else if typ != "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "unsupported type"), http.StatusBadRequest)
		return
	}

	flows, err := GetFlows(q, page, limit)
	if err != nil {
		logger.Error.Printf("get the flow failed. q: %+v, err: %s\n", *q, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	totalFlows, err := CountFlows(q)
	if err != nil {
		logger.Error.Printf("count the flows failed. q: %+v, err: %s\n", *q, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	numOfCurFlows := len(flows)

	flowIDs := make([]int64, 0, numOfCurFlows)
	for _, v := range flows {
		flowIDs = append(flowIDs, v.ID)
	}

	flowMapCounter, err := CountNodes(flowIDs)
	if err != nil {
		logger.Error.Printf("count the nodes failed. ids:%+v, err: %s\n", flowIDs, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	resp := &RespFlowList{}
	resp.Data = make([]*FlowSummary, 0, numOfCurFlows)

	for _, v := range flows {
		f := &FlowSummary{ID: v.ID, Name: v.Name, IntentName: v.IntentName, NumOfNodes: flowMapCounter[v.ID]}
		resp.Data = append(resp.Data, f)
	}

	resp.Page.Current = page
	resp.Page.Limit = limit
	resp.Page.Total = uint64(totalFlows)

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleDeleteFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	affected, err := DeleteFlow(id, enterprise)
	if err != nil {
		logger.Error.Printf("delete the flow failed. id:%d, err: %s\n", id, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "No such flow id"), http.StatusBadRequest)
		return
	}

}

func handleModifyFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	var req reqNewFlow
	err = util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	affected, err := UpdateFlowName(id, enterprise, req.Name)
	if err != nil {
		logger.Error.Printf("update the flow failed. id:%d, err: %s\n", id, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "No such flow id"), http.StatusBadRequest)
		return
	}
}

func handleNodeOrder(w http.ResponseWriter, r *http.Request) {
	idStr := general.ParseID(r)
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var order []string
	err := util.ReadJSON(r, &order)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	if len(order) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "empty nodes"), http.StatusBadRequest)
		return
	}

	err = UpdateNodeOrder(id, order)
	if err != nil {
		logger.Error.Printf("adjust the order of nodes failed. id:%d, err: %s\n", id, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func handleNewNode(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	groupInReq := SentenceGroupInReq{}
	err = util.ReadJSON(r, &groupInReq)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	group := sentenceGroupInReqToSentenceGroup(&groupInReq)
	group.Enterprise = enterprise
	group.Type = typeMapping["call_in"]
	if group.Position == -1 || group.Role == -1 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "bad sentence group"), http.StatusBadRequest)
		return
	}

	err = NewNode(id, group)
	if err != nil {
		if err == ErrNilFlow {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "No such flow id"), http.StatusBadRequest)
		} else {
			logger.Error.Printf("create the node failed. id:%d, err: %s\n", id, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
		return
	}

	resp := sentenceGroupToSentenceGroupInResponse(group)
	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s \n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleModifyIntent(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	var req reqCallInIntent
	err = util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	if typ, ok := callInIntentMap[req.Type]; ok {

		flow := &model.NavFlowUpdate{}
		ignore := 0
		if typ == 1 {
			groupInReq := SentenceGroupInReq{
				Name:      req.IntentName,
				Role:      req.Role,
				Sentences: req.Sentences,
				Type:      req.Type,
			}

			group := sentenceGroupInReqToSentenceGroup(&groupInReq)
			group.Enterprise = enterprise
			group.Type = typeMapping["call_in"]

			if req.IntentName != "" {

				if group.Position == -1 || group.Role == -1 {
					util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "bad sentence group"), http.StatusBadRequest)
					return
				}

				createdGroup, err := CreateSentenceGroup(group)
				if err != nil {
					logger.Error.Printf("error while create sentence in handleCreateSentenceGroup, reason: %s\n", err.Error())
					util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
					return
				}
				flow.IntentLinkID = &createdGroup.ID
				if req.IntentName != "" {
					flow.IntentName = &req.IntentName
				}
			}

		} else {
			ignore = 1
		}
		flow.IgnoreIntent = &ignore

		_, err = UpdateFlow(id, enterprise, flow)
		if err != nil {
			logger.Error.Printf("update flow information failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}

	} else {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "wrong request type"), http.StatusBadRequest)
	}

}

type apiFlowCreateBody struct {
	CreateTime int64  `json:"create_time"`
	FileName   string `json:"file_name"`
}

type apiFlowCreateResp struct {
	UUID string `json:"id"`
}

//handleFlowCreate is the entry tp create the conversation on the fly
func handleFlowCreate(w http.ResponseWriter, r *http.Request) {
	//get the first available bot and its first scenario
	enterprise := requestheader.GetEnterpriseID(r)
	user := requestheader.GetUserID(r)
	//appid := requestheader.GetAppID(r)

	var requestBody apiFlowCreateBody
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	uuid, err := createFlowConversation(enterprise, user, &requestBody)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		if err == ErrNoModels {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, ErrNoModels.Error()), http.StatusBadRequest)
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
		return
	}

	resp := &apiFlowCreateResp{UUID: uuid}
	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

type apiFlowFinish struct {
	FinishTime int64 `json:"finish_time"`
}

func handleFlowFinish(w http.ResponseWriter, r *http.Request) {
	uuid := general.ParseID(r)

	var requestBody apiFlowFinish
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = finishFlowQI(&requestBody, uuid)
	if err != nil {
		if err == ErrEndTimeSmaller {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("Finish the qi flow failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}

		return
	}

}

func handleFlowUpdate(w http.ResponseWriter, r *http.Request, call *model.Call) {
	req, err := extractNewCallReq(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("request error: %v", err))
		return
	}

	if req.RemoteFile == "" {
		err = fmt.Errorf("Remote file path not specified for realtime QI flow, call UUID: %s",
			call.UUID)
		logger.Error.Println(err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()),
			http.StatusBadRequest)
		return
	}

	resp := RealtimeCallResp{
		CallID:     call.ID,
		CallUUID:   call.UUID,
		RemoteFile: req.RemoteFile,
	}

	p, err := json.Marshal(&resp)
	if err != nil {
		logger.Error.Printf("Marshal realtime call resp failed, error: %s",
			err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.IO_ERROR,
			err.Error()), http.StatusBadRequest)
		return
	}

	err = realtimeCallProducer.Produce(p)
	if err != nil {
		logger.Error.Printf("Cannot create realtime call download task: %s",
			err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.IO_ERROR, err.Error()),
			http.StatusInternalServerError)
		return
	}

	err = updateFlowQI(req, call)
	if err != nil {
		logger.Error.Printf("Update qi flow failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

//the speaker in wording in navigation flow only
const (
	WordStaff    = "staff"
	WordCustomer = "customer"
	WordSilence  = "silence"
)

//the speaker in int in navigation flow only
const (
	ChannelSilence = iota
	ChannelStaff
	ChannelCustomer
)

func asrContentToSegment(callID int64, a []model.AsrContent) ([]model.RealSegment, error) {
	num := len(a)
	resp := make([]model.RealSegment, 0, num)
	now := time.Now().Unix()
	for _, v := range a {
		s := model.RealSegment{CallID: callID, StartTime: v.StartTime, EndTime: v.EndTime, Text: v.Text, CreateTime: now}
		switch v.Speaker {
		case WordStaff:
			s.Channel = ChannelStaff
		case WordCustomer:
			s.Channel = ChannelCustomer
		case WordSilence:
			s.Channel = ChannelSilence
		default:
			return nil, errors.New("unknown speaker " + v.Speaker)
		}
		resp = append(resp, s)
	}
	return resp, nil
}

var navCache timecache.TimeCache
var navCallCache timecache.TimeCache

func setUpNavCache() {
	config := &timecache.TCacheConfig{}
	config.SetCollectionDuration(30 * time.Second)
	config.SetCollectionMethod(timecache.OnUpdate)
	navCache.Activate(config)
	navCallCache.Activate(config)
}

//NavMatchedResponse is the returned structure of handleStreaming
type NavMatchedResponse struct {
	NavResult []MatchedFlowNode `json:"nav_result"`
	Sensitive []string          `json:"sensitive"`
}

func handleStreaming(w http.ResponseWriter, r *http.Request) {
	uuid := general.ParseID(r)
	enterprise := requestheader.GetEnterpriseID(r)
	var requestBody []model.AsrContent
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "empty sentence"), http.StatusBadRequest)
		return
	}
	/*
		calls, err := Calls(dbLike.Conn(), model.CallQuery{UUID: []string{uuid}, EnterpriseID: &enterprise})
		if err != nil {
			logger.Error.Printf("get call failed. %s %s. %s\n", enterprise, uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if len(calls) == 0 {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
			return
		}
		call := &calls[0]
	*/

	cacheKey := uuid + enterprise
	val, ok := navCallCache.GetCache(cacheKey)
	call, ok2 := val.(*model.Call)
	if !ok || !ok2 {
		calls, err := Calls(dbLike.Conn(), model.CallQuery{UUID: []string{uuid}, EnterpriseID: &enterprise})
		if err != nil {
			logger.Error.Printf("get call failed. %s %s. %s\n", enterprise, uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if len(calls) == 0 {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
			return
		}
		call = &calls[0]
		navCallCache.SetCache(cacheKey, call)
	}

	predicts, err := getStreamingPredict(call.ID)
	if err != nil {
		logger.Error.Printf("get streaming predicts failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(predicts) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}

	segs, err := asrContentToSegment(call.ID, requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	_, err = segmentDao.NewSegments(dbLike.Conn(), segs)
	if err != nil {
		logger.Error.Printf("insert segments failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	var channelRoles = map[int8]int{
		1: int(call.LeftChanRole),
		2: int(call.RightChanRole),
	}

	segWithSp := make([]*SegmentWithSpeaker, 0, len(segs))
	for _, s := range segs {
		ws := &SegmentWithSpeaker{
			RealSegment: s,
			Speaker:     channelRoles[s.Channel],
		}
		segWithSp = append(segWithSp, ws)
	}

	var settings NavFlowSetting
	err = json.Unmarshal([]byte(predicts[0].Predict), &settings)
	if err != nil {
		logger.Error.Printf("unmarshal predicts failed. %d, %s\n", call.ID, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	matchedInfo, err := streamingMatch(segWithSp, &settings)
	if err != nil {
		logger.Error.Printf("streaming matched failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	/*
		predict, err := json.Marshal(settings)
		if err != nil {
			logger.Error.Printf("marshal setting failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}

		_, err = updateStreamingPredict(call.ID, string(predict))
		if err != nil {
			logger.Error.Printf("update setting %d failed. %s\n", call.ID, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
	*/

	resp := &NavMatchedResponse{NavResult: matchedInfo, Sensitive: make([]string, 0)}

	//pre-check the sensitive word
	sws := make([]string, 0)
	for i := 0; i < len(requestBody); i++ {
		words, err := sensitive.IsSensitive(requestBody[i].Text)
		if err != nil {
			logger.Error.Printf("get sensitive words failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if len(words) > 0 {
			sws = append(sws, words...)
		}
	}

	//sensitive words happened, check the exception
	if len(sws) > 0 {

		//query the previous segments to check exception, which makes the system slow
		allSegs, err := segmentDao.Segments(dbLike.Conn(), model.SegmentQuery{CallID: []int64{call.ID}, Channel: []int8{1, 2}})
		if err != nil {
			logger.Error.Printf("get segements failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		segWithSp = make([]*SegmentWithSpeaker, 0)
		for _, s := range allSegs {
			ws := &SegmentWithSpeaker{
				RealSegment: s,
				Speaker:     channelRoles[s.Channel],
			}
			segWithSp = append(segWithSp, ws)
		}

		credits, err := SensitiveWordsVerificationWithPacked(call.ID, segWithSp, enterprise)
		if err != nil {
			logger.Error.Printf("get sensitive words failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}

		//here simply find the invalid sensitive word then break
		//which is not accuracy
		for _, v := range credits {
			if v.sensitiveWord.Valid == 0 {
				resp.Sensitive = sws
				break
			}
		}
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("write json failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleGetCurCheck(w http.ResponseWriter, r *http.Request) {
	uuid := general.ParseID(r)
	enterprise := requestheader.GetEnterpriseID(r)

	calls, err := Calls(dbLike.Conn(), model.CallQuery{UUID: []string{uuid}, EnterpriseID: &enterprise})
	if err != nil {
		logger.Error.Printf("get call failed. %s %s. %s\n", enterprise, uuid, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(calls) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}
	call := calls[0]

	predicts, err := getStreamingPredict(call.ID)
	if err != nil {
		logger.Error.Printf("get streaming predicts failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(predicts) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}

	var settings NavFlowSetting
	err = json.Unmarshal([]byte(predicts[0].Predict), &settings)
	if err != nil {
		logger.Error.Printf("unmarshal predicts failed. %d, %s\n", call.ID, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, &settings.NavResult)
	if err != nil {
		logger.Error.Printf("write json failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

//WithFlowCallIDEnterpriseCheck checks the call id and its enterprise
func WithFlowCallIDEnterpriseCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enterprise := requestheader.GetEnterpriseID(r)

		uuid := general.ParseID(r)
		val, ok := navCache.GetCache(uuid)
		expect, ok2 := val.(string)
		if !ok || !ok2 {

			calls, err := Calls(dbLike.Conn(), model.CallQuery{UUID: []string{uuid}, EnterpriseID: &enterprise})
			if err != nil {
				logger.Error.Printf("get call failed. %s %s. %s\n", enterprise, uuid, err)
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
				return
			}
			if len(calls) == 0 {
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
				return
			}
			navCache.SetCache(uuid, enterprise)
		} else {
			if expect != enterprise {
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	}

}
