package qi

import (
	autil "emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util"
	"emotibot.com/emotigo/pkg/logger"
	"net/http"
)

type ConversationRuleInReq struct {
	UUID        string   `json:"rule_id"`
	Name        string   `json:"rule_name"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Min         int      `json:"min"`
	Max         int      `json:"max"`
	Type        string   `json:"type"`
	Score       int      `json:"score"`
	Flows       []string `json:"flows"`
}

type ConversationRuleInRes struct {
	UUID        string                         `json:"rule_id"`
	Name        string                         `json:"rule_name,omitempty"`
	Description string                         `json:"description,omitempty"`
	Severity    string                         `json:"severity,omitempty"`
	Min         int                            `json:"min,omitempty"`
	Max         int                            `json:"max,omitempty"`
	Type        string                         `json:"type,omitempty"`
	Score       int                            `json:"score,omitempty"`
	Flows       []model.SimpleConversationFlow `json:"flows,omitempty"`
}

var severityStringToCode map[string]int8 = map[string]int8{
	"normal":   int8(0),
	"critical": int8(1),
}

var severityCodeToString map[int8]string = map[int8]string{
	int8(0): "normal",
	int8(1): "critical",
}

var typeStringToCode map[string]int8 = map[string]int8{
	"positive": int8(1),
	"negative": int8(-1),
}

var typeCodeToString map[int8]string = map[int8]string{
	int8(1):  "positive",
	int8(-1): "negative",
}

func ruleInReqToConversationRule(ruleInReq *ConversationRuleInReq) (rule *model.ConversationRule) {
	rule = &model.ConversationRule{
		Name:        ruleInReq.Name,
		Min:         ruleInReq.Min,
		Max:         ruleInReq.Max,
		Score:       ruleInReq.Score,
		Description: ruleInReq.Description,
	}

	rule.Severity = severityStringToCode[ruleInReq.Severity]
	rule.Type = typeStringToCode[ruleInReq.Type]

	flows := make([]model.SimpleConversationFlow, len(ruleInReq.Flows))
	for idx := range ruleInReq.Flows {
		simpleFlow := model.SimpleConversationFlow{
			UUID: ruleInReq.Flows[idx],
		}
		flows[idx] = simpleFlow
	}
	rule.Flows = flows
	return
}

func conversationRuleToRuleInRes(rule *model.ConversationRule) (ruleInRes *ConversationRuleInRes) {
	ruleInRes = &ConversationRuleInRes{
		UUID:        rule.UUID,
		Name:        rule.Name,
		Severity:    severityCodeToString[rule.Severity],
		Min:         rule.Min,
		Max:         rule.Max,
		Type:        typeCodeToString[rule.Type],
		Score:       rule.Score,
		Flows:       rule.Flows,
		Description: rule.Description,
	}
	return
}

func handleCreateConversationRule(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	ruleInReq := ConversationRuleInReq{}
	err := autil.ReadJSON(r, &ruleInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rule := ruleInReqToConversationRule(&ruleInReq)
	rule.Enterprise = enterprise

	createdRule, err := CreateConversationRule(rule)
	if err != nil {
		logger.Error.Printf("error while create conversation rule in handleCreateConversationRule, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ruleInRes := ConversationRuleInRes{
		UUID: createdRule.UUID,
	}

	autil.WriteJSON(w, ruleInRes)
}

func handleGetConversationRules(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	filter := &model.ConversationRuleFilter{
		Enterprise: enterprise,
		Severity:   -1,
	}

	total, rules, err := GetConversationRulesBy(filter)
	if err != nil {
		logger.Error.Printf("error while get rules in handleGetConversationRules, reason: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Paging *util.Paging
		Data   []ConversationRuleInRes
	}

	rulesInRes := make([]ConversationRuleInRes, len(rules))
	for idx := range rules {
		ruleInRes := conversationRuleToRuleInRes(&rules[idx])
		rulesInRes[idx] = *ruleInRes
	}

	response := Response{
		Paging: &util.Paging{
			Page:  0,
			Total: total,
			Limit: len(rules),
		},
		Data: rulesInRes,
	}

	autil.WriteJSON(w, response)
}

func handleGetConversationRule(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	id := parseID(r)

	filter := &model.ConversationRuleFilter{
		UUID: []string{
			id,
		},
		Enterprise: enterprise,
		Severity:   -1,
	}

	_, rules, err := GetConversationRulesBy(filter)
	if err != nil {
		logger.Error.Printf("error while get rule in handleGetConversationRule, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(rules) == 0 {
		http.NotFound(w, r)
		return
	}

	rule := rules[0]
	ruleInRes := conversationRuleToRuleInRes(&rule)

	autil.WriteJSON(w, ruleInRes)
}

func handleUpdateConversationRule(w http.ResponseWriter, r *http.Request) {

}

func handleDeleteConversationRule(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)

	err := DeleteConversationRule(id)
	if err != nil {
		logger.Error.Printf("error while delete rule in handleDeleteConversationRule, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
