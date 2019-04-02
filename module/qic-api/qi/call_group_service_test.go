package qi

import (
	"database/sql"
	"encoding/json"
	"testing"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
)

var (
	callIDs = []int64{153}
)

var testDB *sql.DB

func newTestDB(t *testing.T) model.DBLike {
	if testDB == nil {
		db, err := sql.Open("mysql", "root:password@tcp(192.168.3.78)/QISYS?parseTime=true&loc=Asia%2FTaipei")
		db.SetMaxIdleConns(0)
		if err != nil {
			t.Fatal("expect db open success but got error: ", err)
		}
		testDB = db
	}
	dbLike = &model.DefaultDBLike{
		DB: testDB,
	}
	return dbLike
}

type cgstMockCreditDao struct {
	mockCreditDao
}

func (m *cgstMockCreditDao) GetCallCredit(conn model.SqlLike, q *model.CreditQuery) ([]*model.SimpleCredit, error) {
	return cgstTestCredits[testCase], nil
}

type cgstMockRuleDaoCredit struct {
	mockRuleDaoCredit
}

func (m *cgstMockRuleDaoCredit) GetBy(filter *model.ConversationRuleFilter, sql model.SqlLike) ([]model.ConversationRule, error) {
	return []model.ConversationRule{
		model.ConversationRule{ID: 1, Method: 1, Score: 1},
		model.ConversationRule{ID: 2, Method: 1, Score: -10},
		model.ConversationRule{ID: 3, Method: -1, Score: 100},
		model.ConversationRule{ID: 4, Method: -1, Score: -1000},
	}, nil
}

type cgstMockCreditCallGroupDao struct {
	model.CreditCallGroupSQLDao
}

func (*cgstMockCreditCallGroupDao) CreateCreditCallGroup(conn model.SqlLike, model *model.CreditCallGroup) (int64, error) {
	return 0, nil
}

func (*cgstMockCreditCallGroupDao) UpdateCreditCallGroup(conn model.SqlLike, query *model.GeneralQuery, data *model.CreditCallGroupUpdateSet) (int64, error) {
	return 0, nil
}

var testCase = "case1"
var callGroupID uint64 = 1
var creatTime int64 = 1500000000
var cgstTestCredits = map[string][]*model.SimpleCredit{
	"case1": []*model.SimpleCredit{
		&model.SimpleCredit{ID: 1, CallID: 1, Type: 0, ParentID: 0, OrgID: 0, Valid: 0, Revise: 0, Score: -809, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 2, CallID: 1, Type: 1, ParentID: 1, OrgID: 1, Valid: -1, Revise: -1, Score: -999, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 3, CallID: 1, Type: 1, ParentID: 1, OrgID: 2, Valid: -1, Revise: -1, Score: 90, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 4, CallID: 1, Type: 10, ParentID: 2, OrgID: 1, Valid: 1, Revise: -1, Score: 1, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 5, CallID: 1, Type: 10, ParentID: 2, OrgID: 2, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 6, CallID: 1, Type: 10, ParentID: 2, OrgID: 3, Valid: 0, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 7, CallID: 1, Type: 10, ParentID: 2, OrgID: 4, Valid: 0, Revise: -1, Score: -1000, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 8, CallID: 1, Type: 10, ParentID: 3, OrgID: 1, Valid: 0, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 9, CallID: 1, Type: 10, ParentID: 3, OrgID: 2, Valid: 0, Revise: -1, Score: -10, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 10, CallID: 1, Type: 10, ParentID: 3, OrgID: 3, Valid: 1, Revise: -1, Score: 100, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 11, CallID: 1, Type: 10, ParentID: 3, OrgID: 4, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},

		&model.SimpleCredit{ID: 12, CallID: 2, Type: 0, ParentID: 0, OrgID: 0, Valid: 0, Revise: 0, Score: 280, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 13, CallID: 2, Type: 1, ParentID: 12, OrgID: 1, Valid: -1, Revise: -1, Score: 90, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 14, CallID: 2, Type: 1, ParentID: 12, OrgID: 2, Valid: -1, Revise: -1, Score: 90, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 15, CallID: 2, Type: 10, ParentID: 13, OrgID: 1, Valid: 0, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 16, CallID: 2, Type: 10, ParentID: 13, OrgID: 2, Valid: 0, Revise: -1, Score: -10, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 17, CallID: 2, Type: 10, ParentID: 13, OrgID: 3, Valid: 1, Revise: -1, Score: 100, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 18, CallID: 2, Type: 10, ParentID: 13, OrgID: 4, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 19, CallID: 2, Type: 10, ParentID: 14, OrgID: 1, Valid: 0, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 20, CallID: 2, Type: 10, ParentID: 14, OrgID: 2, Valid: 0, Revise: -1, Score: -10, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 21, CallID: 2, Type: 10, ParentID: 14, OrgID: 3, Valid: 1, Revise: -1, Score: 100, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 22, CallID: 2, Type: 10, ParentID: 14, OrgID: 4, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},

		&model.SimpleCredit{ID: 23, CallID: 3, Type: 0, ParentID: 0, OrgID: 0, Valid: 0, Revise: 0, Score: 280, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 24, CallID: 3, Type: 1, ParentID: 23, OrgID: 1, Valid: -1, Revise: -1, Score: 90, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 25, CallID: 3, Type: 1, ParentID: 23, OrgID: 2, Valid: -1, Revise: -1, Score: 90, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 26, CallID: 3, Type: 10, ParentID: 24, OrgID: 1, Valid: 0, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 27, CallID: 3, Type: 10, ParentID: 24, OrgID: 2, Valid: 0, Revise: -1, Score: -10, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 28, CallID: 3, Type: 10, ParentID: 24, OrgID: 3, Valid: 1, Revise: -1, Score: 100, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 29, CallID: 3, Type: 10, ParentID: 24, OrgID: 4, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 30, CallID: 3, Type: 10, ParentID: 25, OrgID: 1, Valid: 0, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 31, CallID: 3, Type: 10, ParentID: 25, OrgID: 2, Valid: 0, Revise: -1, Score: -10, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 32, CallID: 3, Type: 10, ParentID: 25, OrgID: 3, Valid: 1, Revise: -1, Score: 100, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 33, CallID: 3, Type: 10, ParentID: 25, OrgID: 4, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
	},
}

func getExpectedCreditCGTree(tCase string) *CallGroupCreditCGTree {
	expectedResult := map[string]*CallGroupCreditCGTree{
		"case1": &CallGroupCreditCGTree{
			Credit: &model.CreditCallGroup{Score: -809},
			RuleGroupMap: map[uint64]*ruleGroupCreditCG{
				1: &ruleGroupCreditCG{
					Credit: &model.CreditCallGroup{Type: 1, OrgID: 1, Valid: -1, Score: -999},
					RuleMap: map[uint64]*ruleCreditCG{
						1: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 1, Valid: 1, Score: 1},
						},
						2: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 2, Valid: 1, Score: 0},
						},
						3: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 3, Valid: 0, Score: 0},
						},
						4: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 4, Valid: 0, Score: -1000},
						},
					},
				},
				2: &ruleGroupCreditCG{
					Credit: &model.CreditCallGroup{Type: 1, OrgID: 1, Valid: -1, Score: 90},
					RuleMap: map[uint64]*ruleCreditCG{
						1: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 1, Valid: 0, Score: 0},
						},
						2: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 2, Valid: 0, Score: -10},
						},
						3: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 3, Valid: 1, Score: 100},
						},
						4: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 4, Valid: 1, Score: 0},
						},
					},
				},
			},
		},
	}
	return expectedResult[tCase]
}

func TestCreateCreditCallGroups(t *testing.T) {
	// dbLike = newTestDB(t)
	dbLike = &test.MockDBLike{}
	creditDao = &cgstMockCreditDao{}
	conversationRuleDao = &cgstMockRuleDaoCredit{}
	creditCallGroupDao = &cgstMockCreditCallGroupDao{}

	for tCase := range cgstTestCredits {
		testCase = tCase
		var expected = getExpectedCreditCGTree(tCase)
		creditTree, ruleIDs, err := GetCallGroupCreditTree(callIDs)
		if err != nil {
			t.Fatalf("expecting no error, but get %s\n", err)
		}
		creditCGTree, err := CreateCreditCallGroups(callGroupID, creditTree, ruleIDs)
		if err != nil {
			t.Fatalf("expecting no error, but get %s\n", err)
		}

		out, _ := json.Marshal(creditCGTree)
		t.Log("cgCreditTree")
		t.Log(string(out))

		if len(expected.RuleGroupMap) != len(creditCGTree.RuleGroupMap) {
			t.Fatalf("expect %d CreditCG of rule group, but get %d\n", len(expected.RuleGroupMap), len(creditCGTree.RuleGroupMap))
		}
		expCredit := expected.Credit
		resultCredit := creditCGTree.Credit
		if expCredit.Score != resultCredit.Score {
			t.Fatalf("expect call group score %d, but get %d\n", expCredit.Score, resultCredit.Score)
		}
		for rgID, rgCreditCG := range creditCGTree.RuleGroupMap {
			rgExp := expected.RuleGroupMap[rgID]
			expCredit := rgExp.Credit
			resultCredit := rgCreditCG.Credit
			if expCredit.Score != resultCredit.Score {
				t.Fatalf("expect rule group score %d, but get %d\n", expCredit.Score, resultCredit.Score)
			}
			if expCredit.Valid != resultCredit.Valid {
				t.Fatalf("expect rule group valid %d, but get %d\n", expCredit.Valid, resultCredit.Valid)
			}
			for rID, rCreditCG := range rgCreditCG.RuleMap {
				rExp := rgExp.RuleMap[rID]
				expCredit := rExp.Credit
				resultCredit := rCreditCG.Credit
				if expCredit.Score != resultCredit.Score {
					t.Fatalf("expect rule score %d, but get %d\n", expCredit.Score, resultCredit.Score)
				}
				if expCredit.Valid != resultCredit.Valid {
					t.Fatalf("expect rule valid %d, but get %d\n", expCredit.Valid, resultCredit.Valid)
				}
			}
		}
	}
}

var getGCTestCase = "case1"

type GetGroupedCallsMockData struct {
	Calls         []model.Call
	CallGroupList []*model.CallGroup
	CallRespList  []CallResp
}

var getGroupedCallsTestData = map[string]*GetGroupedCallsMockData{
	"case1": &GetGroupedCallsMockData{
		Calls: []model.Call{
			model.Call{ID: 1, CallUnixTime: 0},
			model.Call{ID: 2, CallUnixTime: 0},
		},
		CallGroupList: []*model.CallGroup{
			&model.CallGroup{ID: 1, UUID: "1", LastCallID: 1, Calls: []int64{1}},
			&model.CallGroup{ID: 2, UUID: "2", LastCallID: 4, Calls: []int64{2, 3, 4}},
		},
		CallRespList: []CallResp{
			CallResp{CallID: 1, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			CallResp{CallID: 2, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			CallResp{CallID: 3, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			CallResp{CallID: 4, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
		},
	},
}

// func getExpectedGroupedCalls(tCase string) []*GroupedCallsResp {
var getGroupedCallsExpectedResult = map[string][]*GroupedCallsResp{
	"case1": []*GroupedCallsResp{
		&GroupedCallsResp{CallGroupID: 1, CallGroupUUID: "1", IsGroup: false,
			Setting: &CallResp{CallID: 1, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			Calls: []*CallResp{
				&CallResp{CallID: 1, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			}},
		&GroupedCallsResp{CallGroupID: 1, CallGroupUUID: "2", IsGroup: true,
			Setting: &CallResp{CallID: 4, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			Calls: []*CallResp{
				&CallResp{CallID: 2, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
				&CallResp{CallID: 3, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
				&CallResp{CallID: 4, CallUUID: "72cb03687ab44009a1fdb77beb3be874", CallTime: 0, CallLength: 232.56, FileName: "fuckdean45.wav", Status: 2, UploadTime: 1553824445},
			}},
	},
}

type GetGroupedCallsMock struct {
	// callCount          func(delegatee model.SqlLike, query model.CallQuery) (int64, error)
	calls              func(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error)
	callRespsWithTotal func(query model.CallQuery) (responses []CallResp, total int64, err error)
}

var mock = GetGroupedCallsMock{
	// callCount: func(delegatee model.SqlLike, query model.CallQuery) (int64, error) {
	// 	return xx, nil
	// },
	calls: func(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error) {
		return getGroupedCallsTestData[getGCTestCase].Calls, nil
	},
	callRespsWithTotal: func(query model.CallQuery) (responses []CallResp, total int64, err error) {
		return getGroupedCallsTestData[getGCTestCase].CallRespList, 0, nil
	},
}

type cgstMockCallGroupSQLDao struct {
	model.CallGroupSQLDao
}

func (*cgstMockCallGroupSQLDao) GetCallGroups(conn model.SqlLike, query *model.CallGroupQuery) ([]*model.CallGroup, error) {
	return getGroupedCallsTestData[getGCTestCase].CallGroupList, nil
}

func TestGetGroupedCalls(t *testing.T) {
	// dbLike = newTestDB(t)
	dbLike = &test.MockDBLike{}
	// callCount = mock.callCount
	calls = mock.calls
	callGroupDao = &cgstMockCallGroupSQLDao{}
	callRespsWithTotal = mock.callRespsWithTotal

	for tCase := range getGroupedCallsTestData {
		getGCTestCase = tCase
		expGroupedCalls := getGroupedCallsExpectedResult[tCase]
		expTotal := int64(len(expGroupedCalls))

		query := &model.CallQuery{}
		groupedCalls, total, err := GetGroupedCalls(query)
		if err != nil {
			t.Fatalf("expecting no error, but get %s\n", err)
		}
		out, _ := json.Marshal(groupedCalls)
		t.Log("groupedCalls")
		t.Log(string(out))

		if total != expTotal {
			t.Fatalf("expect %d grouped calls, but get %d\n", expTotal, total)
		}
		if len(groupedCalls) != len(expGroupedCalls) {
			t.Fatalf("expect grouped call list with length %d, but get %d\n", len(expGroupedCalls), len(groupedCalls))
		}
		for idx, groupCall := range groupedCalls {
			expGroupCall := expGroupedCalls[idx]
			if groupCall.IsGroup != expGroupCall.IsGroup {
				t.Fatalf("expect IsGroup: %v for groupCall of idx: %d, but get IsGroup: %v\n", expGroupCall.IsGroup, idx, groupCall.IsGroup)
			}
			if len(groupCall.Calls) != len(expGroupCall.Calls) {
				t.Fatalf("expect %d calls for groupCall of idx: %d, but get %d calls\n", len(expGroupCall.Calls), idx, len(groupCall.Calls))
			}
		}
	}
}
