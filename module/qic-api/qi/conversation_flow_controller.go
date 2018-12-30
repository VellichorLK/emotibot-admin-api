package qi

import (
	autil "emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"net/http"
)

type ConversationFlowInReq struct {
	UUID           string   `json:"flow_id"`
	Name           string   `json:"flow_name"`
	Type           string   `json:"type"`
	Expression     string   `json:"expression"`
	SentenceGroups []string `json:"sentence_groups"`
}

type ConversationFlowInRes struct {
	UUID           string                      `json:"flow_id"`
	Name           string                      `json:"flow_name,omitempty"`
	Type           string                      `json:"type,omitempty"`
	Expression     string                      `json:"expression,omitempty"`
	SentenceGroups []model.SimpleSentenceGroup `json:"sentence_groups,omitempty"`
}

func flowInReqToConversationFlow(flowInReq *ConversationFlowInReq, enterprise string) (flow *model.ConversationFlow) {
	flow = &model.ConversationFlow{
		UUID:       flowInReq.UUID,
		Name:       flowInReq.Name,
		Enterprise: enterprise,
		Type:       flowInReq.Type,
		Expression: flowInReq.Expression,
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
	}
}

func handleCreateConversationFlow(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	flowInReq := ConversationFlowInReq{}
	err := autil.ReadJSON(r, &flowInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flow := flowInReqToConversationFlow(&flowInReq, enterprise)
	createdFlow, err := CreateConversationFlow(flow)

	if err != nil {
		logger.Error.Printf("error while create conversation flow in handleCreateConversaionFlow, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flowInRes := conversationfFlowToFlowInRes(createdFlow)

	autil.WriteJSON(w, flowInRes)
}
