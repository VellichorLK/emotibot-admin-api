package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"testing"
)

var mockConversationFlow1 model.ConversationFlow = model.ConversationFlow{
	ID:   1,
	Name: "flow1",
}

var mockConversationFlow2 model.ConversationFlow = model.ConversationFlow{
	ID:   2,
	Name: "flow2",
}

var mockConversationFlows []model.ConversationFlow = []model.ConversationFlow{
	mockConversationFlow1,
	mockConversationFlow2,
}

type mockConversationFlowDao struct{}

func (m *mockConversationFlowDao) Create(flow *model.ConversationFlow, sql model.SqlLike) (*model.ConversationFlow, error) {
	return &mockConversationFlow1, nil
}

func (m *mockConversationFlowDao) CountBy(filter *model.ConversationFlowFilter, sql model.SqlLike) (int64, error) {
	return int64(len(mockConversationFlows)), nil
}

func (m *mockConversationFlowDao) GetBy(filter *model.ConversationFlowFilter, sql model.SqlLike) ([]model.ConversationFlow, error) {
	return mockConversationFlows, nil
}

func (m *mockConversationFlowDao) Update(id string, flow *model.ConversationFlow, sql model.SqlLike) (*model.ConversationFlow, error) {
	return &mockConversationFlow1, nil

}

func (m *mockConversationFlowDao) Delete(id string, sql model.SqlLike) error {
	return nil
}

func restoreConversationFlowTest(origindbLike model.DBLike, originCFDao model.ConversationFlowDao, originSGDao model.SentenceGroupsSqlDao) {
	dbLike = origindbLike
	conversationFlowDao = originCFDao
	sentenceGroupDao = originSGDao
}

func setupConversationFlowMock() (model.DBLike, model.ConversationFlowDao, model.SentenceGroupsSqlDao) {
	originDBLike := dbLike
	mockDBLike := &mockDBLike{}
	dbLike = mockDBLike

	originCFDao := conversationFlowDao
	mockCFDao := &mockConversationFlowDao{}
	conversationFlowDao = mockCFDao

	originSGDao := sentenceGroupDao
	mockDao := &mockSentenceGroupDao{}
	sentenceGroupDao = mockDao

	return originDBLike, originCFDao, originSGDao
}

func TestCreateConversationFlow(t *testing.T) {
	origindbLike, originCFDao, originSGDao := setupConversationFlowMock()
	defer restoreConversationFlowTest(origindbLike, originCFDao, originSGDao)

	createdFlow, err := CreateConversationFlow(&mockConversationFlow1)
	if err != nil {
		t.Error(err)
		return
	}

	if createdFlow.UUID != mockConversationFlow1.UUID {
		t.Errorf("expect uuid: %s, but got: %s", mockConversationFlow1.UUID, createdFlow.UUID)
		return
	}

	if createdFlow.CreateTime != mockConversationFlow1.CreateTime || createdFlow.UpdateTime != mockConversationFlow1.UpdateTime {
		t.Errorf("expect createTime: %d, but got: %d", mockConversationFlow1.CreateTime, createdFlow.CreateTime)
		t.Errorf("expect createTime: %d, but got: %d", mockConversationFlow1.UpdateTime, createdFlow.UpdateTime)
		return
	}
}

func TestSimpleSentenceGroupsOf(t *testing.T) {
	origindbLike, originCFDao, originSGDao := setupConversationFlowMock()
	defer restoreConversationFlowTest(origindbLike, originCFDao, originSGDao)

	groups, err := simpleSentenceGroupsOf(&mockConversationFlow1, sqlConn)
	if err != nil {
		t.Error(err)
		return
	}

	if len(groups) != len(mockSentenceGroups) {
		t.Errorf("expect group number: %d, but got: %d", len(mockSentenceGroups), len(groups))
		return
	}
}

func TestGetConversationFlowsBy(t *testing.T) {
	origindbLike, originCFDao, originSGDao := setupConversationFlowMock()
	defer restoreConversationFlowTest(origindbLike, originCFDao, originSGDao)

	filter := &model.ConversationFlowFilter{}
	total, flows, err := GetConversationFlowsBy(filter)
	if err != nil {
		t.Error(err)
		return
	}

	if total != int64(len(flows)) {
		t.Error("expect total equal to number of flows")
		return
	}

	if total != int64(len(mockConversationFlows)) {
		t.Errorf("expect %d flows, but got: %d", len(mockConversationFlows), total)
		return
	}
}
