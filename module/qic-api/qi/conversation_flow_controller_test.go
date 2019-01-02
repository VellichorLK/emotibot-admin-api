package qi

import (
	"testing"
)

func TestFlowInReqToConversationFlow(t *testing.T) {
	flowInReq := ConversationFlowInReq{
		UUID:       "string",
		Name:       "name",
		Type:       "type",
		Expression: "expression",
		SentenceGroups: []string{
			"1",
			"2",
		},
	}
	enterprise := "enterprise"

	flow := flowInReqToConversationFlow(&flowInReq, enterprise)

	if flowInReq.UUID != flow.UUID || flowInReq.Name != flow.Name || flowInReq.Type != flow.Type || flowInReq.Expression != flow.Expression || len(flowInReq.SentenceGroups) != len(flow.SentenceGroups) {
		t.Errorf("expect flow: %+v, but got: %+v", flowInReq, flow)
		return
	}

	if flow.Enterprise != enterprise {
		t.Errorf("expect enterprise: %s, but got: %s", enterprise, flow.Enterprise)
		return
	}
}
func TestConversationfFlowToFlowInRes(t *testing.T) {
	flowInRes := conversationfFlowToFlowInRes(&mockConversationFlow1)

	if flowInRes.UUID != mockConversationFlow1.UUID || flowInRes.Name != mockConversationFlow1.Name || flowInRes.Type != mockConversationFlow1.Type || flowInRes.Expression != mockConversationFlow1.Expression || len(flowInRes.SentenceGroups) != len(mockConversationFlow1.SentenceGroups) {
		t.Errorf("expect flow: %+v, but got: %+v", mockConversationFlow1, flowInRes)
		return
	}
}
