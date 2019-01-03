package qi

import (
	"database/sql"
	"testing"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var mockSentence1 *model.Sentence = &model.Sentence{}

var mockSentence2 *model.Sentence = &model.Sentence{}

var mockSimpleSentence1 model.SimpleSentence = model.SimpleSentence{}
var mockSimpleSentence2 model.SimpleSentence = model.SimpleSentence{}

var mockSentences []*model.Sentence = []*model.Sentence{
	mockSentence1,
	mockSentence2,
}

var mockSentenceGroup1 model.SentenceGroup = model.SentenceGroup{
	ID:       55688,
	Name:     "mocksg1",
	Role:     0,
	Position: 1,
	Sentences: []model.SimpleSentence{
		mockSimpleSentence1,
		mockSimpleSentence2,
	},
	Enterprise: "test",
}

var mockSentenceGroup2 model.SentenceGroup = model.SentenceGroup{
	ID:       55689,
	Name:     "mocksg2",
	Role:     0,
	Position: 1,
	Sentences: []model.SimpleSentence{
		mockSimpleSentence1,
		mockSimpleSentence2,
	},
	Enterprise: "test2",
}

var mockSentenceGroups []model.SentenceGroup = []model.SentenceGroup{
	mockSentenceGroup1,
	mockSentenceGroup2,
}

type mockSentenceGroupDao struct{}

func (m *mockSentenceGroupDao) Create(group *model.SentenceGroup, sql model.SqlLike) (*model.SentenceGroup, error) {
	return &mockSentenceGroup1, nil
}

func (m *mockSentenceGroupDao) CountBy(filter *model.SentenceGroupFilter, sql model.SqlLike) (total int64, err error) {
	total = 0
	for _, uuid := range filter.UUID {
		if uuid == mockSentenceGroup1.UUID || uuid == mockSentenceGroup2.UUID {
			total += 1
		}
	}
	return
}

func (m *mockSentenceGroupDao) GetBy(filter *model.SentenceGroupFilter, sql model.SqlLike) (groups []model.SentenceGroup, err error) {
	groups = mockSentenceGroups
	return
}

func (m *mockSentenceGroupDao) Update(id string, group *model.SentenceGroup, sql model.SqlLike) (*model.SentenceGroup, error) {
	return nil, nil
}

func (m *mockSentenceGroupDao) Delete(id string, sql model.SqlLike) error {
	return nil
}

type mockSentencesDao struct{}

func (m *mockSentencesDao) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockSentencesDao) Commit(*sql.Tx) error {
	return nil
}

func (m *mockSentencesDao) GetSentences(tx *sql.Tx, q *model.SentenceQuery) ([]*model.Sentence, error) {
	return mockSentences, nil
}

func (m *mockSentencesDao) InsertSentence(tx *sql.Tx, s *model.Sentence) (int64, error) {
	return 0, nil

}

func (m *mockSentencesDao) SoftDeleteSentence(tx *sql.Tx, q *model.SentenceQuery) (int64, error) {
	return 0, nil
}

func (m *mockSentencesDao) CountSentences(tx *sql.Tx, q *model.SentenceQuery) (uint64, error) {
	return 0, nil
}

func (m *mockSentencesDao) InsertSenTagRelation(tx *sql.Tx, s *model.Sentence) error {
	return nil
}

func (m *mockSentencesDao) GetRelSentenceIDByTagIDs(tx *sql.Tx, tagIDs []uint64) (map[uint64][]uint64, error) {
	return nil, nil
}

type mockDBLike struct{}

func (m *mockDBLike) Begin() (*sql.Tx, error) {
	return nil, nil
}
func (m *mockDBLike) ClearTransition(tx *sql.Tx) {
	return
}

func (m *mockDBLike) Commit(tx *sql.Tx) (err error) {
	return
}

func restoreSentenceGroupTest(dbl model.DBLike, dao model.SentenceGroupsSqlDao, sdao model.SentenceDao) {
	dbLike = dbl
	sentenceGroupDao = dao
	sentenceDao = sdao
}

func setupSentenceGroupTestMock() (model.DBLike, model.SentenceGroupsSqlDao, model.SentenceDao) {
	originDBLike := dbLike
	mockDBLike := &mockDBLike{}
	dbLike = mockDBLike

	originSGDao := sentenceGroupDao
	mockDao := &mockSentenceGroupDao{}
	sentenceGroupDao = mockDao

	originSDao := sentenceDao
	mockSDao := &mockSentencesDao{}
	sentenceDao = mockSDao

	return originDBLike, originSGDao, originSDao
}

func TestCreateSentenceGroup(t *testing.T) {
	originDBLike, originSGDao, originSDao := setupSentenceGroupTestMock()
	defer restoreSentenceGroupTest(originDBLike, originSGDao, originSDao)

	created, err := CreateSentenceGroup(&mockSentenceGroup1)
	if err != nil {
		t.Error(err)
		return
	}

	if created.UUID != mockSentenceGroup1.UUID {
		t.Errorf("expect %s, but got %s", mockSentenceGroup1.UUID, created.UUID)
		return
	}
}

func TestGetSentenceGroups(t *testing.T) {
	originDBLike, originSGDao, originSDao := setupSentenceGroupTestMock()
	defer restoreSentenceGroupTest(originDBLike, originSGDao, originSDao)

	filter := &model.SentenceGroupFilter{
		UUID: []string{
			mockSentenceGroup1.UUID,
			mockSentenceGroup2.UUID,
		},
	}
	total, _, err := GetSentenceGroupsBy(filter)
	if err != nil {
		t.Error(err)
		return
	}

	if total != 2 {
		t.Errorf("expect %d, but got: %d", len(mockGroups), total)
		return
	}
}
