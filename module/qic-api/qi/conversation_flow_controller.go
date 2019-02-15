package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

type ConversationFlowInReq struct {
	UUID           string   `json:"flow_id"`
	Name           string   `json:"flow_name"`
	Type           string   `json:"type"`
	Expression     string   `json:"expression"`
	SentenceGroups []string `json:"sentence_groups"`
	Min            *int     `json:"min"`
}

type ConversationFlowInRes struct {
	UUID           string                      `json:"flow_id"`
	Name           string                      `json:"flow_name,omitempty"`
	Type           string                      `json:"type,omitempty"`
	Expression     string                      `json:"expression"`
	SentenceGroups []model.SimpleSentenceGroup `json:"sentence_groups,omitempty"`
	Min            int                         `json:"min"`
	Max            int                         `json:"max"`
}

func flowInReqToConversationFlow(flowInReq *ConversationFlowInReq, enterprise string) (flow *model.ConversationFlow) {
	flow = &model.ConversationFlow{
		UUID:       flowInReq.UUID,
		Name:       flowInReq.Name,
		Enterprise: enterprise,
		Type:       flowInReq.Type,
		Expression: flowInReq.Expression,
	}

	if flowInReq.Min != nil {
		if *flowInReq.Min == 0 {
			flow.Min = 1
		} else {
			flow.Min = *flowInReq.Min
		}
	} else {
		flow.Min = 1
	}

	simpleGroups := make([]model.SimpleSentenceGroup, len(flowInReq.SentenceGroups))
	for idx := range simpleGroups {
		simepleGroup := model.SimpleSentenceGroup{
			UUID: flowInReq.SentenceGroups[idx],
		}
		simpleGroups[idx] = simepleGroup
	}
	flow.SentenceGroups = simpleGroups
	return
}

func conversationfFlowToFlowInRes(flow *model.ConversationFlow) ConversationFlowInRes {
	return ConversationFlowInRes{
		UUID:           flow.UUID,
		Name:           flow.Name,
		Expression:     flow.Expression,
		Type:           flow.Type,
		SentenceGroups: flow.SentenceGroups,
		Min:            flow.Min,
	}
}

func handleCreateConversationFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	flowInReq := ConversationFlowInReq{}
	err := util.ReadJSON(r, &flowInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flow := flowInReqToConversationFlow(&flowInReq, enterprise)
	createdFlow, err := CreateConversationFlow(flow)

	if err != nil {
		logger.Error.Printf("error while create conversation flow in handleCreateConversaionFlow, reason: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flowInRes := conversationfFlowToFlowInRes(createdFlow)

	util.WriteJSON(w, flowInRes)
}

func handleGetConversationFlows(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	filter := &model.ConversationFlowFilter{
		Enterprise: enterprise,
	}

	total, flows, err := GetConversationFlowsBy(filter)
	if err != nil {
		logger.Error.Printf("error while get conversation flows in handleGetConversationFlows, reason: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flowsInRes := make([]ConversationFlowInRes, len(flows))
	for idx, flow := range flows {
		flowInRes := conversationfFlowToFlowInRes(&flow)
		flowsInRes[idx] = flowInRes
	}

	type Response struct {
		Paging *general.Paging         `json:"page"`
		Data   []ConversationFlowInRes `json:"data"`
	}

	response := Response{
		Paging: &general.Paging{
			Total: total,
			Page:  0,
			Limit: len(flows),
		},
		Data: flowsInRes,
	}
	util.WriteJSON(w, response)
}

func handleGetConversationFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	id := parseID(r)

	var deleted int8
	filter := &model.ConversationFlowFilter{
		Enterprise: enterprise,
		UUID: []string{
			id,
		},
		IsDelete: &deleted,
	}

	_, flows, err := GetConversationFlowsBy(filter)
	if err != nil {
		logger.Error.Printf("error while get conversation flow(%s) in handleGetConversationFlow, reason: %s\n", id, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(flows) == 0 {
		http.NotFound(w, r)
		return
	}

	flow := flows[0]
	flowInRes := conversationfFlowToFlowInRes(&flow)

	util.WriteJSON(w, flowInRes)
	return
}

func handleUpdateConversationFlow(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	enterprise := requestheader.GetEnterpriseID(r)

	flowInReq := ConversationFlowInReq{}
	err := util.ReadJSON(r, &flowInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flow := flowInReqToConversationFlow(&flowInReq, enterprise)
	updatedFlow, err := UpdateConversationFlow(id, enterprise, flow)
	if err != nil {
		logger.Error.Printf("error while update conversation flow in handleUpdateConversationFlow, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flowInRes := conversationfFlowToFlowInRes(updatedFlow)

	util.WriteJSON(w, flowInRes)
}

func handleDeleteConversationFlow(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)

	err := DeleteConversationFlow(id)
	if err != nil {
		logger.Error.Printf("error while delete conversation flow in handleDeleteConversationFlow, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
