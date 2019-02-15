package qi

import (
	"testing"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
)

type mockNavDao struct {
}

var mockNewFlowID int64 = 100
var mockAffectedRow int64 = 1

var mockEnterprise = "myenterpirse"
var mockNavID int64 = 1
var mockFlowName = "mockflowname"
var mockFlows = []*model.NavFlow{
	&model.NavFlow{ID: 1, Name: "hello1", IntentName: "intent11"},
	&model.NavFlow{ID: 2, Name: "hello2", IntentName: "intent22"},
	&model.NavFlow{ID: 3, Name: "hello3", IntentName: "intent33"},
}

var mockCountNodes = map[int64]int64{
	1: 1,
	2: 5,
	3: 6,
}

func (m *mockNavDao) NewFlow(conn model.SqlLike, p *model.NavFlow) (int64, error) {
	return mockNewFlowID, nil
}
func (m *mockNavDao) GetFlows(conn model.SqlLike, q *model.NavQuery, l *model.NavLimit) ([]*model.NavFlow, error) {
	return mockFlows, nil
}
func (m *mockNavDao) InsertRelation(conn model.SqlLike, parent int64, child int64) (int64, error) {
	return 0, nil
}
func (m *mockNavDao) DeleteRelation(conn model.SqlLike, parent int64) (int64, error) {
	return 0, nil
}
func (m *mockNavDao) SoftDeleteFlows(conn model.SqlLike, q *model.NavQuery) (int64, error) {
	return mockAffectedRow, nil
}
func (m *mockNavDao) DeleteFlows(conn model.SqlLike, q *model.NavQuery) (int64, error) {
	return mockAffectedRow, nil
}
func (m *mockNavDao) CountFlows(conn model.SqlLike, q *model.NavQuery) (int64, error) {
	return int64(len(mockFlows)), nil
}
func (m *mockNavDao) CountNodes(conn model.SqlLike, navs []int64) (map[int64]int64, error) {
	return mockCountNodes, nil
}
func (m *mockNavDao) GetNodeID(conn model.SqlLike, nav int64) ([]int64, error) {
	return nil, nil
}
func (m *mockNavDao) UpdateFlows(conn model.SqlLike, q *model.NavQuery, d *model.NavFlowUpdate) (int64, error) {
	return mockAffectedRow, nil
}

func setUpMoackNavDao() {
	navDao = &mockNavDao{}
}

func setUpMockNav() {
	setupSentenceGroupTestMock()
	setUpMoackNavDao()
}
func TestNewFlow(t *testing.T) {
	setUpMockNav()
	rb := &reqNewFlow{Name: "hello world", IntentName: "intent1", Type: "intent"}
	id, err := NewFlow(rb, "myenterprise")
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if mockNewFlowID != id {
		t.Fatalf("expecting get id %d, but get %d\n", mockNewFlowID, id)
	}
}

func TestNewNode(t *testing.T) {
	setUpMockNav()

	err := NewNode(1, nil)
	if err == nil {
		t.Fatalf("expecting get error %s, but get nil", ErrNilSentenceGroup)
	}
	d := &model.SentenceGroup{}
	err = NewNode(1, d)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
}

func TestUpdateFlowName(t *testing.T) {
	setUpMockNav()
	affected, err := UpdateFlowName(mockNavID, mockEnterprise, mockFlowName)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if mockAffectedRow != affected {
		t.Fatalf("expecting get affected row %d, but get %d\n", mockAffectedRow, affected)
	}
}

func TestUpdateFlow(t *testing.T) {
	setUpMockNav()
	d := &model.NavFlowUpdate{}
	affected, err := UpdateFlow(mockNavID, mockEnterprise, d)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if mockAffectedRow != affected {
		t.Fatalf("expecting get affected row %d, but get %d\n", mockAffectedRow, affected)
	}
}

func TestDeleteFlow(t *testing.T) {
	setUpMockNav()
	affected, err := DeleteFlow(mockNavID, mockEnterprise)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if mockAffectedRow != affected {
		t.Fatalf("expecting get affected row %d, but get %d\n", mockAffectedRow, affected)
	}
}

func TestFlowSetting(t *testing.T) {
	setUpMockNav()
	resp, err := GetFlowSetting(mockNavID, mockEnterprise)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if resp == nil {
		t.Fatalf("expecting get none nil, but get nil\n")
	}

	if 0 != len(resp.Nodes) {
		t.Fatalf("expecting get zero nodes, but get %d nodes\n", len(resp.Nodes))
	}

	if 0 != len(resp.Sentences) {
		t.Fatalf("expecting get zeron sentences in intent, but get %d sentences\n", len(resp.Sentences))
	}
}

func TestGetFlow(t *testing.T) {
	setUpMockNav()
	flows, err := GetFlows(nil, 0, 10)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if len(mockFlows) != len(flows) {
		t.Fatalf("expecting get %d flows, but get %d \n", len(mockFlows), len(flows))
	}

	for idx, expect := range mockFlows {
		f := flows[idx]
		if expect.ID != f.ID {
			t.Fatalf("expecting id %d at index %d, but get id %d\n", expect.ID, idx, f.ID)
		}
		if expect.IgnoreIntent != f.IgnoreIntent {
			t.Fatalf("expecting ignoreIntent %d at index %d, but get id %d\n", expect.IgnoreIntent, idx, f.IgnoreIntent)
		}
		if expect.IntentLinkID != f.IntentLinkID {
			t.Fatalf("expecting IntentLinkID %d at index %d, but get id %d\n", expect.IntentLinkID, idx, f.IntentLinkID)
		}
	}
}

func TestCountFlows(t *testing.T) {
	setUpMockNav()
	count, err := CountFlows(nil)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if int64(len(mockFlows)) != count {
		t.Fatalf("expecting count flows %d\n, but get %d\n", len(mockFlows), count)
	}
}

func TestCountNodes(t *testing.T) {
	setUpMockNav()
	nodes, err := CountNodes([]int64{1, 2, 3})
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}
	if len(mockCountNodes) != len(nodes) {
		t.Fatalf("expect %d nodes, but get %d\n", len(mockCountNodes), len(nodes))
	}
}
