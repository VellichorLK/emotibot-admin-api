package qi

import (
	"database/sql"
	"testing"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	test "emotibot.com/emotigo/module/qic-api/util/test"
	"bytes"
)

var mockCredits = []*model.SimpleCredit{
	&model.SimpleCredit{ID: 1, Type: 1, ParentID: 0, OrgID: 100, Valid: 1, Revise: -1, Score: 75, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 2, Type: 1, ParentID: 0, OrgID: 102, Valid: 1, Revise: -1, Score: 75, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 3, Type: 10, ParentID: 1, OrgID: 103, Valid: 0, Revise: -1, Score: -5, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 14, Type: 10, ParentID: 1, OrgID: 114, Valid: 1, Revise: -1, Score: 5, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 4, Type: 10, ParentID: 2, OrgID: 104, Valid: 0, Revise: -1, Score: 5, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 5, Type: 20, ParentID: 3, OrgID: 105, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 15, Type: 20, ParentID: 14, OrgID: 115, Valid: 1, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 6, Type: 20, ParentID: 4, OrgID: 106, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 7, Type: 30, ParentID: 5, OrgID: 107, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 16, Type: 30, ParentID: 15, OrgID: 116, Valid: 1, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 8, Type: 30, ParentID: 6, OrgID: 108, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 9, Type: 40, ParentID: 7, OrgID: 109, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 17, Type: 40, ParentID: 16, OrgID: 117, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 10, Type: 40, ParentID: 8, OrgID: 110, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 11, Type: 50, ParentID: 9, OrgID: 111, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 18, Type: 50, ParentID: 17, OrgID: 118, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 12, Type: 50, ParentID: 10, OrgID: 112, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},
	&model.SimpleCredit{ID: 13, Type: 50, ParentID: 10, OrgID: 113, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},
}

var mockMatched = []*model.SegmentMatch{
	&model.SegmentMatch{ID: 1, SegID: 111, TagID: 1, Score: 78, Match: "bitch", MatchedText: "youbitch"},
	&model.SegmentMatch{ID: 2, SegID: 112, TagID: 2, Score: 78, Match: "fuck", MatchedText: "fuckyou"},
	&model.SegmentMatch{ID: 3, SegID: 113, TagID: 3, Score: 78, Match: "hello", MatchedText: "helloworld"},
	&model.SegmentMatch{ID: 4, SegID: 118, TagID: 4, Score: 78, Match: "no", MatchedText: "no comment"},
}

type mockRelationCreditDao struct {
}

func (m *mockRelationCreditDao) GetLevelRelationID(sql model.SqlLike, from int, to int, id []uint64, ignoreNULL bool) ([]map[uint64][]uint64, [][]uint64, error) {
	resp := make([]map[uint64][]uint64, 0, 1)

	senTag := make(map[uint64][]uint64)
	senTag[109] = []uint64{1}
	senTag[110] = []uint64{2, 3}
	senTag[117] = []uint64{4}
	resp = append(resp, senTag)
	return resp, nil, nil

}

type mockCreditDao struct {
}

func (m *mockCreditDao) InsertCredit(conn model.SqlLike, c *model.SimpleCredit) (int64, error) {
	return 0, nil
}
func (m *mockCreditDao) InsertSegmentMatch(conn model.SqlLike, s *model.SegmentMatch) (int64, error) {
	return 0, nil
}
func (m *mockCreditDao) GetCallCredit(conn model.SqlLike, q *model.CreditQuery) ([]*model.SimpleCredit, error) {
	for _, v := range mockCredits {
		v.CallID = q.Calls[0]
	}
	return mockCredits, nil
}
func (m *mockCreditDao) GetSegmentMatch(conn model.SqlLike, q *model.SegmentPredictQuery) ([]*model.SegmentMatch, error) {

	return mockMatched, nil
}

type mockGroupDaoCredit struct {
}

func (m *mockGroupDaoCredit) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockGroupDaoCredit) Commit(tx *sql.Tx) error {
	return nil
}

func (m *mockGroupDaoCredit) ClearTranscation(tx *sql.Tx) {}

func (m *mockGroupDaoCredit) CreateGroup(group *model.GroupWCond, sqlLike model.SqlLike) (*model.GroupWCond, error) {
	return nil, nil
}

func (m *mockGroupDaoCredit) CountGroupsBy(filter *model.GroupFilter, sqlLike model.SqlLike) (int64, error) {
	return 2, nil
}

func (m *mockGroupDaoCredit) GetGroupsBy(filter *model.GroupFilter, sqlLike model.SqlLike) ([]model.GroupWCond, error) {
	return nil, nil
}

func (m *mockGroupDaoCredit) DeleteGroup(id string, sqlLike model.SqlLike) (err error) {
	return
}

func (m *mockGroupDaoCredit) Group(delegatee model.SqlLike, query model.GroupQuery) ([]model.Group, error) {
	return nil, nil
}

func (m *mockGroupDaoCredit) GroupsByCalls(delegatee model.SqlLike, query model.CallQuery) (map[int64][]model.Group, error) {
	return nil, nil
}

func (m *mockGroupDaoCredit) CreateMany(rules []model.GroupWCond, sqlLike model.SqlLike) error {
	return nil
}

func (m *mockGroupDaoCredit) DeleteMany(ruleUUID []string, sql model.SqlLike) error {
	return nil
}

func (m *mockGroupDaoCredit) GetGroupsByRuleID(id []int64, sqlLike model.SqlLike) ([]model.GroupWCond, error) {
	return nil, nil
}

func (m *mockGroupDaoCredit) ExportGroups(sqlLike model.SqlLike) (*bytes.Buffer, error) {
	return nil, nil
}
func (m *mockGroupDaoCredit) ImportGroups(sqlLike model.SqlLike, fileName string) error {
	return nil
}

type mockRuleDaoCredit struct {
}

func (m *mockRuleDaoCredit) Create(rule *model.ConversationRule, sql model.SqlLike) (*model.ConversationRule, error) {
	return nil, nil
}
func (m *mockRuleDaoCredit) CountBy(filter *model.ConversationRuleFilter, sql model.SqlLike) (int64, error) {
	return 0, nil
}
func (m *mockRuleDaoCredit) GetBy(filter *model.ConversationRuleFilter, sql model.SqlLike) ([]model.ConversationRule, error) {
	return nil, nil
}
func (m *mockRuleDaoCredit) Delete(id string, sql model.SqlLike) error {
	return nil
}

func (m *mockRuleDaoCredit) CreateMany(rules []model.ConversationRule, sqlLike model.SqlLike) error {
	return nil
}

func (m *mockRuleDaoCredit) DeleteMany(ruleUUID []string, sql model.SqlLike) error {
	return nil
}

func (m *mockRuleDaoCredit) GetByFlowID(flowID []int64, sql model.SqlLike) ([]model.ConversationRule, error) {
	return nil, nil
}

type mockCFDaoCredit struct {
}

func (m *mockCFDaoCredit) Create(flow *model.ConversationFlow, sql model.SqlLike) (*model.ConversationFlow, error) {
	return nil, nil
}
func (m *mockCFDaoCredit) CountBy(filter *model.ConversationFlowFilter, sql model.SqlLike) (int64, error) {
	return 0, nil
}
func (m *mockCFDaoCredit) GetBy(filter *model.ConversationFlowFilter, sql model.SqlLike) ([]model.ConversationFlow, error) {
	return nil, nil
}
func (m *mockCFDaoCredit) Update(id string, flow *model.ConversationFlow, sql model.SqlLike) (*model.ConversationFlow, error) {
	return nil, nil
}
func (m *mockCFDaoCredit) Delete(id string, sql model.SqlLike) error {
	return nil
}

func (m *mockCFDaoCredit) CreateMany(flows []model.ConversationFlow, sql model.SqlLike) error {
	return nil
}

func (m *mockCFDaoCredit) DeleteMany(id []string, sql model.SqlLike) error {
	return nil
}

func (m *mockCFDaoCredit) GetBySentenceGroupID(SGID []int64, sql model.SqlLike) ([]model.ConversationFlow, error) {
	return nil, nil
}

type mockSenGrpCredit struct {
}

func (m *mockSenGrpCredit) Create(group *model.SentenceGroup, sql model.SqlLike) (*model.SentenceGroup, error) {
	return nil, nil
}
func (m *mockSenGrpCredit) CountBy(filter *model.SentenceGroupFilter, sql model.SqlLike) (int64, error) {
	return 0, nil
}
func (m *mockSenGrpCredit) GetBy(filter *model.SentenceGroupFilter, sql model.SqlLike) ([]model.SentenceGroup, error) {
	return nil, nil
}
func (m *mockSenGrpCredit) Update(id string, group *model.SentenceGroup, sql model.SqlLike) (*model.SentenceGroup, error) {
	return nil, nil
}
func (m *mockSenGrpCredit) Delete(id string, sqllike model.SqlLike) error {
	return nil
}
func (m *mockSenGrpCredit) GetNewBy(id []int64, filter *model.SentenceGroupFilter, sql model.SqlLike) ([]model.SentenceGroup, error) {
	return nil, nil
}

func (m *mockSenGrpCredit) CreateMany(sgs []model.SentenceGroup, sql model.SqlLike) error {
	return nil
}

func (m *mockSenGrpCredit) DeleteMany(id []string, sql model.SqlLike) error {
	return nil
}

func (m *mockSenGrpCredit) GetBySentenceID(id []int64, sql model.SqlLike) ([]model.SentenceGroup, error) {
	return nil, nil
}

type mockTagSQLDaoCredit struct {
	data            map[uint64]model.Tag
	uuidData        map[string]model.Tag
	numByEnterprise map[string]int
	enterprises     []string
	uuid            []string
}

func (m *mockTagSQLDaoCredit) Tags(tx model.SqlLike, q model.TagQuery) ([]model.Tag, error) {
	return nil, nil
}

func (m *mockTagSQLDaoCredit) NewTags(tx model.SqlLike, tags []model.Tag) ([]model.Tag, error) {
	return nil, nil
}

func (m *mockTagSQLDaoCredit) DeleteTags(tx model.SqlLike, query model.TagQuery) (int64, error) {
	return 0, nil
}
func (m *mockTagSQLDaoCredit) CountTags(tx model.SqlLike, query model.TagQuery) (uint, error) {
	return 0, nil
}

type mockSentenceSQLDaoCredit struct {
	data            map[uint64]*model.Sentence
	uuidData        map[string]*model.Sentence
	numByEnterprise map[string]int
	enterprises     []string
	uuid            []string
}

func (m *mockSentenceSQLDaoCredit) Begin() (*sql.Tx, error) {
	return nil, nil

}

func (m *mockSentenceSQLDaoCredit) Commit(tx *sql.Tx) error {
	return nil
}

func (m *mockSentenceSQLDaoCredit) MoveCategories(x model.SqlLike, q *model.SentenceQuery, category uint64) (int64, error) {
	return 0, nil
}

func (m *mockSentenceSQLDaoCredit) InsertSenTagRelation(tx model.SqlLike, s *model.Sentence) error {
	return nil
}

func (m *mockSentenceSQLDaoCredit) GetRelSentenceIDByTagIDs(tx model.SqlLike, tagIDs []uint64) (map[uint64][]uint64, error) {
	return nil, nil
}

func (m *mockSentenceSQLDaoCredit) GetSentences(tx model.SqlLike, q *model.SentenceQuery) ([]*model.Sentence, error) {
	return nil, nil
}
func (m *mockSentenceSQLDaoCredit) InsertSentence(tx model.SqlLike, s *model.Sentence) (int64, error) {
	return 0, nil
}
func (m *mockSentenceSQLDaoCredit) SoftDeleteSentence(tx model.SqlLike, q *model.SentenceQuery) (int64, error) {
	return 0, nil
}
func (m *mockSentenceSQLDaoCredit) CountSentences(tx model.SqlLike, q *model.SentenceQuery) (uint64, error) {
	return 0, nil
}

func (m *mockTagSQLDaoCredit) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockSentenceSQLDaoCredit) InsertSentences(tx model.SqlLike, sentences []model.Sentence) error {
	return nil
}

/*
var mockCredits = []*model.SimpleCredit{
	&model.SimpleCredit{ID: 1, Type: 1, ParentID: 0, OrgID: 100, Valid: 1, Revise: -1, Score: 75, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 2, Type: 1, ParentID: 0, OrgID: 102, Valid: 1, Revise: -1, Score: 75, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 3, Type: 10, ParentID: 1, OrgID: 103, Valid: 0, Revise: -1, Score: -5, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 14, Type: 10, ParentID: 1, OrgID: 114, Valid: 1, Revise: -1, Score: 5, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 4, Type: 10, ParentID: 2, OrgID: 104, Valid: 0, Revise: -1, Score: 5, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 5, Type: 20, ParentID: 3, OrgID: 105, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 15, Type: 20, ParentID: 14, OrgID: 115, Valid: 1, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 6, Type: 20, ParentID: 4, OrgID: 106, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 7, Type: 30, ParentID: 5, OrgID: 107, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 16, Type: 30, ParentID: 15, OrgID: 116, Valid: 1, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 8, Type: 30, ParentID: 6, OrgID: 108, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 9, Type: 40, ParentID: 7, OrgID: 109, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 17, Type: 40, ParentID: 16, OrgID: 117, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 10, Type: 40, ParentID: 8, OrgID: 110, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},

	&model.SimpleCredit{ID: 11, Type: 50, ParentID: 9, OrgID: 111, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 18, Type: 50, ParentID: 17, OrgID: 118, Valid: 0, Revise: -1, Score: 0, CreateTime: 100, UpdateTime: 100},
	&model.SimpleCredit{ID: 12, Type: 50, ParentID: 10, OrgID: 112, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},
	&model.SimpleCredit{ID: 13, Type: 50, ParentID: 10, OrgID: 113, Valid: 0, Revise: -1, Score: 0, CreateTime: 200, UpdateTime: 200},
}

*/

func getExpect() []*HistoryCredit {
	var expect = []*HistoryCredit{
		&HistoryCredit{CreateTime: 200},
		&HistoryCredit{CreateTime: 100},
	}
	r := mockCredits[0]
	rGrpCredit1 := &RuleGrpCredit{ID: r.OrgID, Score: r.Score}
	r = mockCredits[1]
	rGrpCredit2 := &RuleGrpCredit{ID: r.OrgID, Score: r.Score}

	r = mockCredits[2]
	rCredit1 := &RuleCredit{ID: r.OrgID, Valid: validMap[r.Valid], Score: r.Score}
	r = mockCredits[3]
	rCredit2 := &RuleCredit{ID: r.OrgID, Valid: validMap[r.Valid], Score: r.Score}
	r = mockCredits[4]
	rCredit3 := &RuleCredit{ID: r.OrgID, Valid: validMap[r.Valid], Score: r.Score}

	r = mockCredits[5]
	cfCredit1 := &ConversationFlowCredit{ID: r.OrgID, Valid: validMap[r.Valid]}
	r = mockCredits[6]
	cfCredit2 := &ConversationFlowCredit{ID: r.OrgID, Valid: validMap[r.Valid]}
	r = mockCredits[7]
	cfCredit3 := &ConversationFlowCredit{ID: r.OrgID, Valid: validMap[r.Valid]}

	r = mockCredits[8]
	senGrpCredit1 := &SentenceGrpCredit{ID: r.OrgID, Valid: validMap[r.Valid]}
	r = mockCredits[9]
	senGrpCredit2 := &SentenceGrpCredit{ID: r.OrgID, Valid: validMap[r.Valid]}
	r = mockCredits[10]
	senGrpCredit3 := &SentenceGrpCredit{ID: r.OrgID, Valid: validMap[r.Valid]}

	r = mockCredits[11]
	senCredit1 := &SentenceCredit{ID: r.OrgID, Valid: validMap[r.Valid]}
	r = mockCredits[12]
	senCredit2 := &SentenceCredit{ID: r.OrgID, Valid: validMap[r.Valid]}
	r = mockCredits[13]
	senCredit3 := &SentenceCredit{ID: r.OrgID, Valid: validMap[r.Valid]}

	senCredit1.MatchedSegments = append(senCredit1.MatchedSegments, mockMatched[0])
	senCredit2.MatchedSegments = append(senCredit2.MatchedSegments, mockMatched[3])
	senCredit3.MatchedSegments = append(senCredit3.MatchedSegments, mockMatched[1])
	senCredit3.MatchedSegments = append(senCredit3.MatchedSegments, mockMatched[2])

	senGrpCredit1.Sentences = append(senGrpCredit1.Sentences, senCredit1)
	senGrpCredit2.Sentences = append(senGrpCredit2.Sentences, senCredit2)
	senGrpCredit3.Sentences = append(senGrpCredit3.Sentences, senCredit3)

	cfCredit1.SentenceGrps = append(cfCredit1.SentenceGrps, senGrpCredit1)
	cfCredit2.SentenceGrps = append(cfCredit2.SentenceGrps, senGrpCredit2)
	cfCredit3.SentenceGrps = append(cfCredit3.SentenceGrps, senGrpCredit3)

	rCredit1.CFs = append(rCredit1.CFs, cfCredit1)
	rCredit2.CFs = append(rCredit2.CFs, cfCredit2)
	rCredit3.CFs = append(rCredit3.CFs, cfCredit3)

	rGrpCredit1.Rules = append(rGrpCredit1.Rules, rCredit1)
	rGrpCredit1.Rules = append(rGrpCredit1.Rules, rCredit2)
	rGrpCredit2.Rules = append(rGrpCredit2.Rules, rCredit3)

	expect[0].Credit = append(expect[0].Credit, rGrpCredit2)
	expect[1].Credit = append(expect[1].Credit, rGrpCredit1)

	return expect
}

func TestRetreiveCredit(t *testing.T) {
	//creditDao = &mockGroupDaoCredit{}
	creditDao = &mockCreditDao{}
	serviceDAO = &mockGroupDaoCredit{}
	conversationRuleDao = &mockRuleDaoCredit{}
	conversationFlowDao = &mockCFDaoCredit{}
	sentenceGroupDao = &mockSenGrpCredit{}
	sentenceDao = &mockSentenceSQLDaoCredit{}
	relationDao = &mockRelationCreditDao{}
	tagDao = &mockTagSQLDaoCredit{}
	dbLike = &test.MockDBLike{}
	credits, err := RetrieveCredit(1234)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	var expect = getExpect()

	if len(expect) != len(credits) {
		t.Fatalf("expect %d history, but get %d\n", len(expect), len(credits))
	}
	for idx, credit := range credits {
		if expect[idx].CreateTime != credit.CreateTime {
			t.Fatalf("expect create time %d, but get %d\n", expect[idx].CreateTime, credit.CreateTime)
		}
		if len(expect[idx].Credit) != len(credit.Credit) {
			t.Fatalf("expect %d credits in %d'th history, but get %d\n", len(expect[idx].Credit), idx, len(credit.Credit))
		}

		expectGrp := expect[idx]
		for rgIdx, ruleGrps := range credit.Credit {
			expectruleGrps := expectGrp.Credit[rgIdx]
			if expectruleGrps.ID != ruleGrps.ID {
				t.Fatalf("expect get rule groups id %d at %d'th, but get %d\n", expectruleGrps.ID, rgIdx, ruleGrps.ID)
			}
			if expectruleGrps.Score != ruleGrps.Score {
				t.Fatalf("expect rule groups id %d get %d score,  but get %d\n", expectruleGrps.ID, expectruleGrps.Score, ruleGrps.Score)
			}
			if len(expectruleGrps.Rules) != len(ruleGrps.Rules) {
				t.Fatalf("expect rule groups id %d get %d rules,  but get %d\n", expectruleGrps.ID, len(expectruleGrps.Rules), len(ruleGrps.Rules))
			}

			for rIdx, rule := range ruleGrps.Rules {
				expectRule := expectruleGrps.Rules[rIdx]
				if expectRule.ID != rule.ID {
					t.Fatalf("expect ruile %d at %d'th, but get %d\n", expectRule.ID, rIdx, rule.ID)
				}
				if expectRule.Score != rule.Score {
					t.Fatalf("expect ruile %d with %d score, but get %d\n", expectRule.ID, expectRule.Score, rule.Score)
				}
				if expectRule.Valid != rule.Valid {
					t.Fatalf("expect ruile %d with %t valid, but get %t\n", expectRule.ID, expectRule.Valid, rule.Valid)
				}
				if len(expectRule.CFs) != len(rule.CFs) {
					t.Fatalf("expect ruile %d with %d conversation flow, but get %d\n", expectRule.ID, len(expectRule.CFs), len(rule.CFs))
				}
				for cfIdx, cf := range rule.CFs {
					expectCF := expectRule.CFs[cfIdx]
					if expectCF.ID != cf.ID {
						t.Fatalf("expect cf %d at %d'th, but get %d\n", expectCF.ID, cfIdx, cf.ID)
					}
					if expectCF.Valid != cf.Valid {
						t.Fatalf("expect cf %d with %t valid, but get %t\n", expectCF.ID, expectCF.Valid, cf.Valid)
					}
					if len(expectCF.SentenceGrps) != len(cf.SentenceGrps) {
						t.Fatalf("expect cf %d with %d sentence group, but get %d\n", expectCF.ID, len(expectCF.SentenceGrps), len(cf.SentenceGrps))
					}
					for senGrpIdx, senGrp := range cf.SentenceGrps {
						expectSenGrp := expectCF.SentenceGrps[senGrpIdx]
						if expectSenGrp.ID != senGrp.ID {
							t.Fatalf("expect sentence group %d at %d'th, but get %d\n", expectSenGrp.ID, senGrpIdx, senGrp.ID)
						}
						if expectSenGrp.Valid != senGrp.Valid {
							t.Fatalf("expect sentence group %d with %t valid, but get %t\n", expectSenGrp.ID, expectSenGrp.Valid, senGrp.Valid)
						}
						if len(expectSenGrp.Sentences) != len(senGrp.Sentences) {
							t.Fatalf("expect sentence group %d with %d sentence, but get %d\n", expectSenGrp.ID, len(expectSenGrp.Sentences), len(senGrp.Sentences))
						}
						for senIdx, sen := range senGrp.Sentences {
							expectSen := expectSenGrp.Sentences[senIdx]
							if expectSen.ID != sen.ID {
								t.Fatalf("expect sentence  %d at %d'th, but get %d\n", expectSen.ID, senIdx, sen.ID)
							}
							if expectSen.Valid != sen.Valid {
								t.Fatalf("expect sentence %d with %t valid, but get %t\n", expectSen.ID, expectSen.Valid, sen.Valid)
							}
							if len(expectSen.MatchedSegments) != len(sen.MatchedSegments) {
								t.Fatalf("expect sentence %d with %d matched, but get %d\n", expectSen.ID, len(expectSen.MatchedSegments), len(sen.MatchedSegments))
							}
							for segIdx, seg := range sen.MatchedSegments {
								expectSeg := expectSen.MatchedSegments[segIdx]
								if expectSeg.ID != seg.ID {
									t.Fatalf("expect segment id %d, but get %d\n", expectSeg.ID, seg.ID)
								}
								if expectSeg.SegID != seg.SegID {
									t.Fatalf("expect segment segment_id %d, but get %d\n", expectSeg.SegID, seg.SegID)
								}
								if expectSeg.Match != seg.Match {
									t.Fatalf("expect segment match %s, but get %s\n", expectSeg.Match, seg.Match)
								}
								if expectSeg.MatchedText != seg.MatchedText {
									t.Fatalf("expect segment match text %s, but get %s\n", expectSeg.MatchedText, seg.MatchedText)
								}
								if expectSeg.Score != seg.Score {
									t.Fatalf("expect segment score %d, but get %d\n", expectSeg.Score, seg.Score)
								}
								if expectSeg.TagID != seg.TagID {
									t.Fatalf("expect segment tag %d, but get %d\n", expectSeg.TagID, seg.TagID)
								}
							}
						}
					}
				}
			}

		}

	}
}
