package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"testing"
)

var mockConversationRule1 model.ConversationRule = model.ConversationRule{
	Name:   "rule1",
	Method: 1,
	Score:  100,
	Min:    1,
	Max:    5,
}

var mockConversationRule2 model.ConversationRule = model.ConversationRule{
	Name:   "rule2",
	Method: 1,
	Score:  50,
	Min:    1,
	Max:    5,
}

var mockConversationRules []model.ConversationRule = []model.ConversationRule{
	mockConversationRule1,
	mockConversationRule2,
}

type mockConversationRuleDao struct{}

func (m *mockConversationRuleDao) Create(rule *model.ConversationRule, sql model.SqlLike) (*model.ConversationRule, error) {
	return &mockConversationRule1, nil
}

func (m *mockConversationRuleDao) CountBy(filter *model.ConversationRuleFilter, sql model.SqlLike) (int64, error) {
	return int64(len(mockConversationRules)), nil
}
func (m *mockConversationRuleDao) GetBy(filter *model.ConversationRuleFilter, sql model.SqlLike) ([]model.ConversationRule, error) {
	return mockConversationRules, nil
}

func (m *mockConversationRuleDao) Delete(id string, sql model.SqlLike) error {
	return nil
}

func setupConversationRuleMock() (model.DBLike, model.ConversationRuleDao, model.ConversationFlowDao) {
	originDBLike := dbLike
	mockDBLike := &mockDBLike{}
	dbLike = mockDBLike

	originCRDao := conversationRuleDao
	mockCRDao := &mockConversationRuleDao{}
	conversationRuleDao = mockCRDao

	originCFDao := conversationFlowDao
	mockCFDao := &mockConversationFlowDao{}
	conversationFlowDao = mockCFDao

	return originDBLike, originCRDao, originCFDao
}

func restoreConversationRuleTest(origindbLike model.DBLike, originCRDao model.ConversationRuleDao, originCFDao model.ConversationFlowDao) {
	dbLike = origindbLike
	conversationRuleDao = originCRDao
	conversationFlowDao = originCFDao
}

func TestCreateConversationRule(t *testing.T) {
	originDBLike, originCRDao, originCFDao := setupConversationRuleMock()
	defer restoreConversationRuleTest(originDBLike, originCRDao, originCFDao)

	createdRule, err := CreateConversationRule(&mockConversationRule1)
	if err != nil {
		t.Error(err)
		return
	}

	if createdRule.UUID == "" {
		t.Error("should create uuid but got empty string")
		return
	}

	if createdRule.UUID != mockConversationRule1.UUID {
		t.Errorf("expect rule UUID: %s, but got: %s", mockConversationRule1.UUID, createdRule.UUID)
		return
	}
}
