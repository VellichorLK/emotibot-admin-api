package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"testing"
)

func TestRuleInReqToConversationRule(t *testing.T) {
	ruleInReq := &ConversationRuleInReq{
		UUID:     "id",
		Name:     "name",
		Severity: "critical",
		Method:   "negative",
		Score:    1,
	}

	rule := ruleInReqToConversationRule(ruleInReq)

	if rule.Name != ruleInReq.Name || rule.Score != ruleInReq.Score {
		t.Errorf("expect rule: %+v, but got: %+v", ruleInReq, rule)
		return
	}

	if rule.Severity != 1 {
		t.Error("transform rule severity fail")
		return
	}

	if rule.Method != -1 {
		t.Error("transform rule type fail")
		return
	}
}

func TestConversationRuleToRuleInRes(t *testing.T) {
	rule := &model.ConversationRule{
		UUID:     "id",
		Name:     "name",
		Severity: 1,
		Method:   -1,
		Score:    1,
	}

	ruleInRes := conversationRuleToRuleInRes(rule)

	if ruleInRes.Severity != "critical" {
		t.Error("transform rule severity fail")
		return
	}

	if ruleInRes.Method != "negative" {
		t.Error("transform rule type fail")
		return
	}

	if rule.UUID != ruleInRes.UUID || rule.Name != ruleInRes.Name || rule.Score != ruleInRes.Score {
		t.Errorf("expect rule: %+v, but got: %+v", rule, ruleInRes)
		return
	}
}
