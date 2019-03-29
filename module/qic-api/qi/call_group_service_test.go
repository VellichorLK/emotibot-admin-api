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
	// callIDMap := make(map[uint64]bool)
	// for _, callID := range q.Calls {
	// 	callIDMap[callID] = true
	// }
	// creditList := []*model.SimpleCredit{}
	// for _, v := range cgstCredits {
	// 	if _, ok := callIDMap[v.CallID]; ok {
	// 		creditList = append(creditList, v)
	// 	}
	// }
	// return creditList, nil
	return cgstTestCredits[testCase], nil
}

type cgstMockRuleDaoCredit struct {
	mockRuleDaoCredit
}

func (m *cgstMockRuleDaoCredit) GetBy(filter *model.ConversationRuleFilter, sql model.SqlLike) ([]model.ConversationRule, error) {
	return []model.ConversationRule{
		model.ConversationRule{ID: 1, Method: -1, Score: -5},
		model.ConversationRule{ID: 2, Method: -1, Score: -5},
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
		&model.SimpleCredit{ID: 1, CallID: 1, Type: 0, ParentID: 0, OrgID: 0, Valid: 0, Revise: 0, Score: 75, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 2, CallID: 1, Type: 1, ParentID: 1, OrgID: 1, Valid: -1, Revise: -1, Score: -10, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 3, CallID: 1, Type: 10, ParentID: 2, OrgID: 1, Valid: 0, Revise: -1, Score: -5, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 4, CallID: 1, Type: 10, ParentID: 2, OrgID: 2, Valid: 0, Revise: -1, Score: -5, CreateTime: creatTime, UpdateTime: creatTime},

		&model.SimpleCredit{ID: 5, CallID: 2, Type: 0, ParentID: 0, OrgID: 0, Valid: 0, Revise: 0, Score: 75, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 6, CallID: 2, Type: 1, ParentID: 5, OrgID: 1, Valid: -1, Revise: -1, Score: -15, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 7, CallID: 1, Type: 10, ParentID: 6, OrgID: 1, Valid: 1, Revise: -1, Score: 0, CreateTime: creatTime, UpdateTime: creatTime},
		&model.SimpleCredit{ID: 8, CallID: 1, Type: 10, ParentID: 6, OrgID: 2, Valid: 0, Revise: -1, Score: -5, CreateTime: creatTime, UpdateTime: creatTime},
	},
}

func getExpectedCreditCGTree(tCase string) *CallGroupCreditCGTree {
	expectedResult := map[string]*CallGroupCreditCGTree{
		"case1": &CallGroupCreditCGTree{
			Credit: &model.CreditCallGroup{Score: 90},
			RuleGroupMap: map[uint64]*ruleGroupCreditCG{
				1: &ruleGroupCreditCG{
					Credit: &model.CreditCallGroup{Type: 1, OrgID: 1, Valid: -1, Score: -10},
					RuleMap: map[uint64]*ruleCreditCG{
						1: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 1, Valid: 0, Score: -5},
						},
						2: &ruleCreditCG{
							Credit: &model.CreditCallGroup{Type: 10, OrgID: 2, Valid: 0, Score: -5},
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
