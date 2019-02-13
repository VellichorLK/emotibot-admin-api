package qi

import (
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"

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
	IntentName string                    `json:"intent_name"`
	Role       string                    `json:"role"`
	Type       string                    `json:"type"`
	Sentences  []model.SimpleSentence    `json:"sentences"`
	Nodes      []SentenceGroupInResponse `json:"nodes"`
}

var callInIntentMap = map[string]int{
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
	resp := &respDetailFlow{}

	if d.SentenceGroup.Sentences == nil {
		resp.Sentences = []model.SimpleSentence{}
	} else {
		resp.Sentences = d.SentenceGroup.Sentences
	}

	if d.IntentLinkID == 0 || d.IgnoreIntent == 1 {
		resp.Type = callInIntentCodeMap[0]
		resp.Sentences = []model.SimpleSentence{}
	} else {
		resp.Type = callInIntentCodeMap[1]
		resp.IntentName = d.IntentName
		resp.Role = roleCodeMap[d.SentenceGroup.Role]
	}

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
	Name string `json:"name"`
}

func handleNewFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	var req reqNewFlow
	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	id, err := NewFlow(req.Name, enterprise)
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
		logger.Error.Printf("get the flow failed. %q:s, err: %s\n", *q, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	totalFlows, err := CountFlows(q)
	if err != nil {
		logger.Error.Printf("count the flows failed. %q:s, err: %s\n", *q, err)
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
		logger.Error.Printf("create the node failed. id:%d, err: %s\n", id, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
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
