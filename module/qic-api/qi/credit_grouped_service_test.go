package qi

import (
	"testing"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
)

var mockCallGroups = []*model.CallGroup{
	&model.CallGroup{
		ID: 1, UUID: "callgroupuuid", IsDelete: 0, CallGroupConditionID: 1, Enterprise: "enterpriseuuid",
		LastCallID: 1, LastCallTime: creatTime, CreateTime: creatTime, UpdateTime: creatTime, Calls: []int64{1, 2, 3}},
}

var mockCreditCGs = []*model.CreditCallGroup{
	&model.CreditCallGroup{ID: 1, CallGroupID: 1, Type: 0, ParentID: 0, OrgID: 0, Valid: 0, Revise: 0, Score: 105, CreateTime: creatTime, UpdateTime: creatTime, CallID: 0},
	&model.CreditCallGroup{ID: 2, CallGroupID: 1, Type: 1, ParentID: 1, OrgID: 100, Valid: -1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime, CallID: 0},
	&model.CreditCallGroup{ID: 3, CallGroupID: 1, Type: 1, ParentID: 1, OrgID: 102, Valid: -1, Revise: -1, Score: 5, CreateTime: creatTime, UpdateTime: creatTime, CallID: 0},

	&model.CreditCallGroup{ID: 4, CallGroupID: 1, Type: 10, ParentID: 2, OrgID: 103, Valid: 0, Revise: -1, Score: -5, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 5, CallGroupID: 1, Type: 10, ParentID: 2, OrgID: 114, Valid: 1, Revise: -1, Score: 5, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 6, CallGroupID: 1, Type: 10, ParentID: 3, OrgID: 104, Valid: 0, Revise: -1, Score: 5, CreateTime: 200, UpdateTime: 200, CallID: 0},

	&model.CreditCallGroup{ID: 7, CallGroupID: 1, Type: 20, ParentID: 4, OrgID: 105, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 8, CallGroupID: 1, Type: 20, ParentID: 5, OrgID: 115, Valid: 1, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 9, CallGroupID: 1, Type: 20, ParentID: 6, OrgID: 106, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200, CallID: 0},

	&model.CreditCallGroup{ID: 10, CallGroupID: 1, Type: 30, ParentID: 7, OrgID: 107, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 11, CallGroupID: 1, Type: 30, ParentID: 8, OrgID: 116, Valid: 1, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 12, CallGroupID: 1, Type: 30, ParentID: 9, OrgID: 108, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200, CallID: 0},

	&model.CreditCallGroup{ID: 13, CallGroupID: 1, Type: 40, ParentID: 10, OrgID: 109, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 14, CallGroupID: 1, Type: 40, ParentID: 11, OrgID: 117, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 15, CallGroupID: 1, Type: 40, ParentID: 12, OrgID: 110, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200, CallID: 0},

	&model.CreditCallGroup{ID: 16, CallGroupID: 1, Type: 50, ParentID: 13, OrgID: 111, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 17, CallGroupID: 1, Type: 50, ParentID: 14, OrgID: 118, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100, CallID: 0},
	&model.CreditCallGroup{ID: 18, CallGroupID: 1, Type: 50, ParentID: 15, OrgID: 112, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200, CallID: 0},
	&model.CreditCallGroup{ID: 19, CallGroupID: 1, Type: 50, ParentID: 15, OrgID: 113, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200, CallID: 0},
}

type mockCallGroupSQLDao struct {
	model.CallGroupSQLDao
}

func (*mockCallGroupSQLDao) GetCallGroups(conn model.SqlLike, query *model.CallGroupQuery) ([]*model.CallGroup, error) {
	return mockCallGroups, nil
}

type mockCreditCallGroupSQLDao struct {
	model.CreditCallGroupSQLDao
}

func (c *mockCreditCallGroupSQLDao) GetCreditCallGroups(conn model.SqlLike, query *model.CreditCallGroupQuery) ([]*model.CreditCallGroup, error) {
	return mockCreditCGs, nil
}

func TestRetrieveGroupedCredit(t *testing.T) {
	dbLike = &test.MockDBLike{}
	callGroupDao = &mockCallGroupSQLDao{}
	creditDao = &mockCreditDao{}
	creditCallGroupDao = &mockCreditCallGroupSQLDao{}

	serviceDAO = &mockGroupDaoCredit{}
	conversationRuleDao = &mockRuleDaoCredit{}
	conversationFlowDao = &mockCFDaoCredit{}
	sentenceGroupDao = &mockSenGrpCredit{}
	sentenceDao = &mockSentenceSQLDaoCredit{}
	relationDao = &mockRelationCreditDao{}
	tagDao = &mockTagSQLDaoCredit{}
	segmentDao = &mockCreditSegmentDao{}

	uuid := "callgroupuuid"
	historyCredits, err := RetrieveGroupedCredit(uuid)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	// out, _ := json.Marshal(historyCredits)
	// t.Log("historyCredits")
	// t.Log(string(out))

	if len(historyCredits) == 0 {
		t.Fatalf("expect %d history, but get %d\n", 1, len(historyCredits))
	}
}
