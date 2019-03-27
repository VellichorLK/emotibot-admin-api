package model

import (
	"encoding/csv"
	"os"
	"testing"
)

func readMockRules(t *testing.T) []ConversationRule {
	f, err := os.Open("./testdata/seed/Rule.csv")
	if err != nil {
		t.Fatal("open file failed, ", err)
	}
	cr := csv.NewReader(f)
	records, err := cr.ReadAll()
	if err != nil {
		t.Fatal("read csv failed, ", err)
	}
	var rules []ConversationRule
	for _, rec := range records[1:] {
		var rule ConversationRule
		Binding(&rule, rec)
		rules = append(rules, rule)
	}
	return rules
}
