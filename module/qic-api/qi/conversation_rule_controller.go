package qi

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

type Completeness struct {
	RuleCompleted       bool `json:"isRuleComplete"`
	HasDescription      bool `json:"hasDescription"`
	HasConversationFlow bool `json:"hasConversationFlow"`
	SetenceCompleted    bool `json:"sentenceComplete"`
}

type ConversationRuleInReq struct {
	UUID        string   `json:"rule_id"`
	Name        string   `json:"rule_name"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Min         int      `json:"min"`
	Max         int      `json:"max"`
	Method      string   `json:"method"`
	Score       int      `json:"score"`
	Flows       []string `json:"flows"`
}

type ConversationRuleInRes struct {
	UUID         string                         `json:"rule_id"`
	Name         string                         `json:"rule_name,omitempty"`
	Description  string                         `json:"description,omitempty"`
	Severity     string                         `json:"severity,omitempty"`
	Min          int                            `json:"min,omitempty"`
	Max          int                            `json:"max,omitempty"`
	Method       string                         `json:"method,omitempty"`
	Score        int                            `json:"score"`
	Flows        []model.SimpleConversationFlow `json:"flows"`
	Completeness *Completeness                  `json:"completeness"`
	Expression   string                         `json:"expression"`
}

var severityStringToCode map[string]int8 = map[string]int8{
	"normal":   int8(0),
	"critical": int8(1),
}

var severityCodeToString map[int8]string = map[int8]string{
	int8(0): "normal",
	int8(1): "critical",
}

var methodStringToCode map[string]int8 = map[string]int8{
	"positive": int8(1),
	"negative": int8(-1),
}

var methodCodeToString map[int8]string = map[int8]string{
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
	rule.Method = methodStringToCode[ruleInReq.Method]

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
		Method:      methodCodeToString[rule.Method],
		Score:       rule.Score,
		Flows:       rule.Flows,
		Description: rule.Description,
		Completeness: &Completeness{
			RuleCompleted:       rule.Completeness.RuleCompleted != int8(0),
			HasDescription:      rule.Completeness.HasDescription,
			HasConversationFlow: rule.Completeness.HasConversationFlow,
			SetenceCompleted:    rule.Completeness.SetenceCompleted != int8(0),
		},
	}
	return
}

func handleCreateConversationRule(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	ruleInReq := ConversationRuleInReq{}
	err := util.ReadJSON(r, &ruleInReq)
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

	util.WriteJSON(w, ruleInRes)
}

func intergrityToRespCompletness(valids []LevelVaild) []*model.RuleCompleteness {
	validRuleCompleteness := &model.RuleCompleteness{RuleCompleted: 1, HasDescription: true, HasConversationFlow: true, SetenceCompleted: 1}
	invalidRuleCompleteness := &model.RuleCompleteness{RuleCompleted: 0, HasDescription: true, HasConversationFlow: false, SetenceCompleted: 0}

	RuleNotCompleteness := &model.RuleCompleteness{RuleCompleted: 0, HasDescription: true, HasConversationFlow: true, SetenceCompleted: 1}
	hasNoConversationFlow := &model.RuleCompleteness{RuleCompleted: 0, HasDescription: true, HasConversationFlow: false, SetenceCompleted: 0}
	sentenceNoCompletness := &model.RuleCompleteness{RuleCompleted: 0, HasDescription: true, HasConversationFlow: true, SetenceCompleted: 0}
	resp := make([]*model.RuleCompleteness, 0, len(valids))

	for _, v := range valids {
		var thisCompleteness model.RuleCompleteness
		if v.Valid {
			thisCompleteness = *validRuleCompleteness
		} else {

			var checker *model.RuleCompleteness

			for _, info := range v.InValidInfo {
				switch info.InValidLevel {
				case LevRuleGroup:
					checker = invalidRuleCompleteness
				case LevRule:
					checker = hasNoConversationFlow
				case LevConversation:
					checker = sentenceNoCompletness
				case LevSenGroup:
					checker = sentenceNoCompletness
				case LevSentence:
					if checker == nil {
						checker = RuleNotCompleteness
					}
				default:
				}
			}
			thisCompleteness = *checker
		}
		resp = append(resp, &thisCompleteness)
	}
	return resp
}

func addIntegrityInfo(rules []model.ConversationRule) error {
	numOfRules := len(rules)
	if numOfRules > 0 {
		ruleIDs := make([]uint64, 0, numOfRules)
		for _, v := range rules {
			ruleIDs = append(ruleIDs, uint64(v.ID))
		}
		ruleValid, err := CheckIntegrity(LevRule, ruleIDs)
		if err != nil || len(ruleValid) != numOfRules {
			logger.Error.Printf("get rule integrity failed. %s. %d, %d\n", err, len(ruleValid), numOfRules)
			return fmt.Errorf("get rule integrity failed. %s", err)

		}

		checkValid := intergrityToRespCompletness(ruleValid)
		for i, v := range checkValid {
			rules[i].Completeness = v
			if rules[i].Description == "" {
				rules[i].Completeness.HasDescription = false
			}
		}

	}
	return nil
}

func handleGetConversationRules(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("parse request failed, %v", err))
		return
	}

	filter := &model.ConversationRuleFilter{
		Enterprise: enterprise,
		Severity:   -1,
		Paging:     &model.Pagination{Limit: limit, Page: page},
	}

	total, rules, err := GetConversationRulesBy(filter)
	if err != nil {
		logger.Error.Printf("error while get rules in handleGetConversationRules, reason: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = addIntegrityInfo(rules)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Paging *general.Paging         `json:"paging"`
		Data   []ConversationRuleInRes `json:"data"`
	}

	rulesInRes := make([]ConversationRuleInRes, len(rules))
	for idx := range rules {
		ruleInRes := conversationRuleToRuleInRes(&rules[idx])
		rulesInRes[idx] = *ruleInRes
	}

	response := Response{
		Paging: &general.Paging{
			Page:  page,
			Total: total,
			Limit: limit,
		},
		Data: rulesInRes,
	}

	util.WriteJSON(w, response)
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

	err = addIntegrityInfo(rules)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	rule := rules[0]
	ruleInRes := conversationRuleToRuleInRes(&rule)

	util.WriteJSON(w, ruleInRes)
}

func handleUpdateConversationRule(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	id := parseID(r)

	ruleInReq := ConversationRuleInReq{}
	err := util.ReadJSON(r, &ruleInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rule := ruleInReqToConversationRule(&ruleInReq)
	rule.Enterprise = enterprise
	rule.UUID = id

	updatedRule, err := UpdateConversationRule(id, rule)
	if err != nil {
		logger.Error.Printf("error while update rule in handleUpdateConversationRule, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if updatedRule == nil {
		http.NotFound(w, r)
		return
	}
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
