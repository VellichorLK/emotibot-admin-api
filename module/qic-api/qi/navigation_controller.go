package qi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/module/qic-api/util/timecache"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
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

func handleFlowList(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	isDelete := 0
	q := &model.NavQuery{Enterprise: &enterprise, IsDelete: &isDelete}

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
		logger.Error.Printf("%s\n", err)
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
	}
}

type apiFlowFinish struct {
	FinishTime int64 `json:"finish_time"`
}

func handleFlowFinish(w http.ResponseWriter, r *http.Request) {
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	var requestBody apiFlowFinish
	err = util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	//FIXME nil
	err = finishFlowQI(&requestBody, id, nil)
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

//the speaker in wording in navigation flow only
const (
	WordHost    = "host"
	WordGuest   = "guest"
	WordSilence = "silence"
)

//the speaker in int in navigation flow only
const (
	ChannelSilence = iota
	ChannelHost
	ChannelGuest
)

func asrContentToSegment(callID int64, a []model.AsrContent) ([]model.RealSegment, error) {
	num := len(a)
	resp := make([]model.RealSegment, 0, num)
	now := time.Now().Unix()
	for _, v := range a {
		s := model.RealSegment{CallID: callID, StartTime: v.StartTime, EndTime: v.EndTime, Text: v.Text, CreateTime: now}
		switch v.Speaker {
		case WordHost:
			s.Channel = ChannelHost
		case WordGuest:
			s.Channel = ChannelGuest
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

func setUpNavCache() {
	config := &timecache.TCacheConfig{}
	config.SetCollectionDuration(30 * time.Second)
	config.SetCollectionMethod(timecache.OnUpdate)
	navCache.Activate(config)
}

func handleStreaming(w http.ResponseWriter, r *http.Request) {
	idStr := general.ParseID(r)
	id, _ := strconv.ParseInt(idStr, 10, 64)

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

	segs, err := asrContentToSegment(id, requestBody)
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

	/*
		//insert the segment
		err = insertSegmentByUUID(uuid, asrContents)
		if err != nil {
			if err == ErrSpeaker {
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
			} else {
				logger.Error.Printf("%s\n", err)
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			}
			return
		}
		enterprise := requestheader.GetEnterpriseID(r)
		resp, err := getCurrentQIFlowResult(w, enterprise, uuid)
		if err != nil {
			return
		}

		callID, err := getIDByUUID(uuid)
		if err != nil {
			logger.Error.Printf("%s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if callID == 0 {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id is found"), http.StatusBadRequest)
			return
		}

		_, err = UpdateFlowResult(callID, resp)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}

		resp.Sensitive = make([]string, 0)

		for i := 0; i < len(requestBody); i++ {

			words, err := sensitive.IsSensitive(requestBody[i].Text)
			if err != nil {
				logger.Error.Printf("get sensitive words failed. %s\n", err)
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
				return
			}

			resp.Sensitive = append(resp.Sensitive, words...)
		}

		err = util.WriteJSON(w, resp)
		if err != nil {
			logger.Error.Printf("%s\n", err)
		}

	*/
}

//WithFlowCallIDEnterpriseCheck checks the call id and its enterprise
func WithFlowCallIDEnterpriseCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enterprise := requestheader.GetEnterpriseID(r)

		idStr := general.ParseID(r)
		val, ok := navCache.GetCache(idStr)
		expect, ok2 := val.(string)
		if !ok || !ok2 {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
				return
			}

			_, err = Call(id, enterprise)
			if err != nil {
				if err == ErrNotFound {
					util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
				} else {
					logger.Error.Printf("get call failed. %s %d. %s\n", enterprise, id, err)
					util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
				}
				return
			}
			navCache.SetCache(idStr, enterprise)
		} else {
			if expect != enterprise {
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	}

}
