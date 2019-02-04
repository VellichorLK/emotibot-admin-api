package qi

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	uuid "github.com/satori/go.uuid"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var mockSentenceDao *mockSentenceSQLDao
var mockTagDao *mockTagSQLDao

func sentenceMockDataSetup() {
	mockSentenceDao = &mockSentenceSQLDao{}

	enterprises := []string{"mycompany", "mycompany2"}
	fakeuuid := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		uuid, err := uuid.NewV4()
		if err != nil {
			fmt.Printf("generate uuid faild. %s\n", err)
			os.Exit(-1)
		}
		uuidStr := hex.EncodeToString(uuid[:])
		fakeuuid = append(fakeuuid, uuidStr)
	}

	mockSentenceDao.data = make(map[uint64]*model.Sentence)
	mockSentenceDao.uuidData = make(map[string]*model.Sentence)
	mockSentenceDao.enterprises = enterprises
	mockSentenceDao.uuid = fakeuuid

	s1 := &model.Sentence{ID: 1, IsDelete: 0, Name: "test1", Enterprise: enterprises[0],
		UUID: fakeuuid[0], CreateTime: 199811, UpdateTime: 1029011, TagIDs: []uint64{1, 2, 3}}

	s2 := &model.Sentence{ID: 2, IsDelete: 0, Name: "test2", Enterprise: enterprises[0],
		UUID: fakeuuid[1], CreateTime: 199813, UpdateTime: 1029013, TagIDs: []uint64{2, 3}}

	s3 := &model.Sentence{ID: 3, IsDelete: 0, Name: "test3", Enterprise: enterprises[1],
		UUID: fakeuuid[2], CreateTime: 199811, UpdateTime: 1029011, TagIDs: []uint64{4, 5}}

	s4 := &model.Sentence{ID: 4, IsDelete: 0, Name: "test4", Enterprise: enterprises[1],
		UUID: fakeuuid[3], CreateTime: 199813, UpdateTime: 1029013, TagIDs: []uint64{5, 6}}

	mockSentenceDao.data[1] = s1
	mockSentenceDao.data[2] = s2
	mockSentenceDao.data[3] = s3
	mockSentenceDao.data[4] = s4

	mockSentenceDao.uuidData[s1.UUID] = s1
	mockSentenceDao.uuidData[s2.UUID] = s2
	mockSentenceDao.uuidData[s3.UUID] = s3
	mockSentenceDao.uuidData[s4.UUID] = s4

	mockSentenceDao.numByEnterprise = make(map[string]int)
	for _, v := range mockSentenceDao.data {
		mockSentenceDao.numByEnterprise[v.Enterprise]++
	}
	sentenceDao = mockSentenceDao

	mockTagDao = &mockTagSQLDao{}
	mockTagDao.data = make(map[uint64]model.Tag)
	mockTagDao.uuidData = make(map[string]model.Tag)
	mockTagDao.enterprises = enterprises
	mockTagDao.uuid = fakeuuid
	t1 := model.Tag{ID: 1, UUID: fakeuuid[0], Name: "tag1", Enterprise: enterprises[0], Typ: 1}
	t2 := model.Tag{ID: 2, UUID: fakeuuid[1], Name: "tag2", Enterprise: enterprises[0], Typ: 1}
	t3 := model.Tag{ID: 3, UUID: fakeuuid[2], Name: "tag3", Enterprise: enterprises[0], Typ: 1}
	t4 := model.Tag{ID: 4, UUID: fakeuuid[3], Name: "tag4", Enterprise: enterprises[1], Typ: 1}
	t5 := model.Tag{ID: 5, UUID: fakeuuid[4], Name: "tag5", Enterprise: enterprises[1], Typ: 1}
	t6 := model.Tag{ID: 6, UUID: fakeuuid[5], Name: "tag6", Enterprise: enterprises[1], Typ: 1}

	mockTagDao.data[t1.ID] = t1
	mockTagDao.data[t2.ID] = t2
	mockTagDao.data[t3.ID] = t3
	mockTagDao.data[t4.ID] = t4
	mockTagDao.data[t5.ID] = t5
	mockTagDao.data[t6.ID] = t6

	mockTagDao.uuidData[t1.UUID] = t1
	mockTagDao.uuidData[t2.UUID] = t2
	mockTagDao.uuidData[t3.UUID] = t3
	mockTagDao.uuidData[t4.UUID] = t4
	mockTagDao.uuidData[t5.UUID] = t5
	mockTagDao.uuidData[t6.UUID] = t6

	mockTagDao.numByEnterprise = make(map[string]int)
	for _, v := range mockTagDao.data {
		mockTagDao.numByEnterprise[v.Enterprise]++
	}

	tagDao = mockTagDao

	mockDBLike := &test.MockDBLike{}
	dbLike = mockDBLike
}

type mockTagSQLDao struct {
	data            map[uint64]model.Tag
	uuidData        map[string]model.Tag
	numByEnterprise map[string]int
	enterprises     []string
	uuid            []string
}

func (m *mockTagSQLDao) Tags(tx *sql.Tx, q model.TagQuery) ([]model.Tag, error) {
	tags := make([]model.Tag, 0)
	for _, id := range q.ID {
		if v, ok := mockTagDao.data[id]; ok {
			if q.Enterprise != nil {
				if *q.Enterprise == v.Enterprise {
					tags = append(tags, v)
				}
			} else {
				tags = append(tags, v)
			}
		}
	}

	for _, id := range q.UUID {
		if v, ok := mockTagDao.uuidData[id]; ok {
			if q.Enterprise != nil {
				if *q.Enterprise == v.Enterprise {
					tags = append(tags, v)
				}
			} else {
				tags = append(tags, v)
			}
		}
	}

	return tags, nil
}

func (m *mockTagSQLDao) NewTags(tx *sql.Tx, tags []model.Tag) ([]model.Tag, error) {
	return nil, nil
}

func (m *mockTagSQLDao) DeleteTags(tx *sql.Tx, query model.TagQuery) (int64, error) {
	return 0, nil
}
func (m *mockTagSQLDao) CountTags(tx *sql.Tx, query model.TagQuery) (uint, error) {
	return 0, nil
}

type mockSentenceSQLDao struct {
	data            map[uint64]*model.Sentence
	uuidData        map[string]*model.Sentence
	numByEnterprise map[string]int
	enterprises     []string
	uuid            []string
}

func (m *mockSentenceSQLDao) Begin() (*sql.Tx, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Printf("sqlmock new failed. %s\n", err)
		os.Exit(-1)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	return db.Begin()

}

func (m *mockSentenceSQLDao) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

func (m *mockSentenceSQLDao) MoveCategories(x model.SqlLike, q *model.SentenceQuery, category uint64) (int64, error) {
	return 0, nil
}

func (m *mockSentenceSQLDao) InsertSenTagRelation(tx model.SqlLike, s *model.Sentence) error {
	return nil
}

func (m *mockSentenceSQLDao) GetRelSentenceIDByTagIDs(tx model.SqlLike, tagIDs []uint64) (map[uint64][]uint64, error) {
	return nil, nil
}

func (m *mockSentenceSQLDao) GetSentences(tx model.SqlLike, q *model.SentenceQuery) ([]*model.Sentence, error) {
	if q == nil {
		return nil, nil
	}
	sentences := make([]*model.Sentence, 0)
	data := make(map[uint64]*model.Sentence)

	if q.ID == nil && q.UUID == nil {
		if q.Enterprise != nil {
			for _, v := range mockSentenceDao.data {
				if v.Enterprise == *q.Enterprise {
					data[v.ID] = v
				}
			}
		}
	} else {

		for _, id := range q.ID {
			if v, ok := mockSentenceDao.data[id]; ok {
				if q.Enterprise != nil {
					if *q.Enterprise == v.Enterprise {
						if q.IsDelete != nil {
							if *q.IsDelete == v.IsDelete {
								data[v.ID] = v
							}
						} else {
							data[v.ID] = v
						}
					}
				} else {
					if q.IsDelete != nil {
						if *q.IsDelete == v.IsDelete {
							data[v.ID] = v
						}
					} else {
						data[v.ID] = v
					}
				}

			}
		}

		for _, id := range q.UUID {
			if v, ok := mockSentenceDao.uuidData[id]; ok {
				if q.Enterprise != nil {
					if *q.Enterprise == v.Enterprise {
						if q.IsDelete != nil {
							if *q.IsDelete == v.IsDelete {
								data[v.ID] = v
							}
						} else {
							data[v.ID] = v
						}
					}
				} else {
					if q.IsDelete != nil {
						if *q.IsDelete == v.IsDelete {
							data[v.ID] = v
						}
					} else {
						data[v.ID] = v
					}
				}
			}
		}
	}

	for _, v := range data {
		sentences = append(sentences, v)
	}
	return sentences, nil

}
func (m *mockSentenceSQLDao) InsertSentence(tx model.SqlLike, s *model.Sentence) (int64, error) {
	if s == nil {
		return 0, nil
	}
	id := uint64(len(mockSentenceDao.data) + 1)
	s.ID = id
	mockSentenceDao.data[id] = s
	mockSentenceDao.uuidData[s.UUID] = s
	return int64(id), nil
}
func (m *mockSentenceSQLDao) SoftDeleteSentence(tx model.SqlLike, q *model.SentenceQuery) (int64, error) {
	if q == nil {
		return 0, nil
	}
	var count int64
	for _, id := range q.ID {
		if v, ok := mockSentenceDao.data[id]; ok {
			if q.Enterprise != nil && *q.Enterprise == v.Enterprise {
				v.IsDelete = 1
				count++
			}
		}
	}
	for _, id := range q.UUID {
		if v, ok := mockSentenceDao.uuidData[id]; ok {
			if q.Enterprise != nil {
				if *q.Enterprise == v.Enterprise {
					if q.IsDelete != nil {
						if *q.IsDelete == v.IsDelete {
							v.IsDelete = 1
							count++
						}
					} else {
						v.IsDelete = 1
						count++
					}

				}
			} else {
				if q.IsDelete != nil {
					if *q.IsDelete == v.IsDelete {
						v.IsDelete = 1
						count++
					}
				} else {
					v.IsDelete = 1
					count++
				}
			}

		}
	}
	return count, nil
}
func (m *mockSentenceSQLDao) CountSentences(tx model.SqlLike, q *model.SentenceQuery) (uint64, error) {
	if q == nil {
		return 0, nil
	}
	data := make(map[uint64]*model.Sentence)

	if q.ID == nil && q.UUID == nil {
		if q.Enterprise != nil {
			for _, v := range mockSentenceDao.data {
				if v.Enterprise == *q.Enterprise {
					data[v.ID] = v
				}
			}
		}
	} else {

		for _, id := range q.ID {
			if v, ok := mockSentenceDao.data[id]; ok {
				if q.Enterprise != nil {
					if *q.Enterprise == v.Enterprise {
						if q.IsDelete != nil {
							if *q.IsDelete == v.IsDelete {
								data[v.ID] = v
							}
						} else {
							data[v.ID] = v
						}
					}
				} else {
					if q.IsDelete != nil {
						if *q.IsDelete == v.IsDelete {
							data[v.ID] = v
						}
					} else {
						data[v.ID] = v
					}
				}

			}
		}

		for _, id := range q.UUID {
			if v, ok := mockSentenceDao.uuidData[id]; ok {
				if q.Enterprise != nil {
					if *q.Enterprise == v.Enterprise {
						if q.IsDelete != nil {
							if *q.IsDelete == v.IsDelete {
								data[v.ID] = v
							}
						} else {
							data[v.ID] = v
						}
					}
				} else {
					if q.IsDelete != nil {
						if *q.IsDelete == v.IsDelete {
							data[v.ID] = v
						}
					} else {
						data[v.ID] = v
					}
				}

			}
		}
	}

	return uint64(len(data)), nil
}

func (m *mockTagSQLDao) Begin() (*sql.Tx, error) {
	return nil, nil
}
func TestCheckSentenceAuth(t *testing.T) {
	sentenceMockDataSetup()
	valid, _ := CheckSentenceAuth([]string{mockSentenceDao.uuid[0]}, mockSentenceDao.enterprises[0])
	if !valid {
		t.Error("case1 expecting get true, but get false\n")
	}

	valid, _ = CheckSentenceAuth([]string{mockSentenceDao.uuid[0], mockSentenceDao.uuid[2]}, mockSentenceDao.enterprises[0])
	if valid {
		t.Error("case2 expecting get false, but get true\n")
	}

	valid, _ = CheckSentenceAuth([]string{mockSentenceDao.uuid[0], mockSentenceDao.uuid[1]}, mockSentenceDao.enterprises[0])
	if !valid {
		t.Error("case3 expecting get true, but get false\n")
	}

	valid, _ = CheckSentenceAuth([]string{mockSentenceDao.uuid[3], mockSentenceDao.uuid[2]}, mockSentenceDao.enterprises[1])
	if !valid {
		t.Error("case4 expecting get true, but get false\n")
	}
}

func TestSoftDeleteSentence(t *testing.T) {
	sentenceMockDataSetup()
	uuid := mockSentenceDao.uuid[0]
	affected, _ := SoftDeleteSentence(uuid, mockSentenceDao.enterprises[0])
	if affected != 1 {
		t.Errorf("case1 expecting get 1, but get %d\n", affected)
	}

	v := mockSentenceDao.uuidData[uuid]
	if v.IsDelete != 1 {
		t.Errorf("case1 expecting Isdelete 1, but get %d", v.IsDelete)
	}

	affected, _ = SoftDeleteSentence(uuid, mockSentenceDao.enterprises[1])
	if affected != 0 {
		t.Errorf("case2 expecting get 0, but get %d\n", affected)
	}

	v = mockSentenceDao.uuidData[uuid]
	if v.IsDelete != 1 {
		t.Errorf("case2 expecting Isdelete 1, but get %d\n", v.IsDelete)
	}
}

func TestUpdateSentence(t *testing.T) {
	t.Skip("any level higher than sentence need to mock now, skip it for a test refractor")
	sentenceMockDataSetup()
	id := uint64(1)
	uuid := mockSentenceDao.uuid[0]
	name := "myupdatename"
	enterprise := mockSentenceDao.enterprises[0]
	tagUUID := []string{"3"}

	newID, _ := UpdateSentence(uuid, name, enterprise, tagUUID)

	if v, ok := mockSentenceDao.data[uint64(newID)]; ok {
		if v.Name != name {
			t.Errorf("case 1 expecting name %s, but get %s\n", name, v.Name)
		}
	} else {
		t.Errorf("case 1 epecting has %d id data, but get none\n", newID)
	}

	if v, ok := mockSentenceDao.data[id]; ok {
		if v.IsDelete != 1 {
			t.Errorf("case 1 expecting IsDelete 1, but get %d\n", v.IsDelete)
		}
	}

}

func TestNewSentence(t *testing.T) {
	sentenceMockDataSetup()
	name := "mynewname"
	enterprise := mockTagDao.enterprises[0]
	tagUUID := []string{mockTagDao.uuid[2]}
	d, _ := NewSentence(enterprise, 0, name, tagUUID)
	if d.Name != name {
		t.Errorf("expecting name:%s, but get %s\n", name, d.Name)
	}

	if v, ok := mockSentenceDao.uuidData[d.UUID]; ok {
		if v.Name != name {
			t.Errorf("expecting name:%s, but get %s\n", name, v.Name)
		}
		if v.Enterprise != enterprise {
			t.Errorf("expecting enterprise:%s, but get %s\n", enterprise, v.Enterprise)
		}
	} else {
		t.Errorf("expecting get %s data, but get none\n", d.UUID)
	}

}

func TestGetSentence(t *testing.T) {
	sentenceMockDataSetup()
	uuid := mockSentenceDao.uuid[0]
	enterprise := mockSentenceDao.enterprises[0]

	d, _ := GetSentence(uuid, enterprise)
	if d == nil {
		t.Errorf("expecting get data, but get none\n")
	}
	if d.UUID != uuid {
		t.Errorf("expecting get data %s, but get %s\n", uuid, d.UUID)
	}
}

func TestGetSentenceList(t *testing.T) {
	sentenceMockDataSetup()
	enterprise := mockSentenceDao.enterprises[0]
	total, d, _ := GetSentenceList(enterprise, 1, 100, nil, nil)
	if total != uint64(len(d)) {
		t.Errorf("expecting total(%d) == get data(%d), but not equal\n", total, len(d))
	}
	if mockSentenceDao.numByEnterprise[enterprise] != int(total) {
		t.Errorf("expecting get %d records, but get %d\n", mockSentenceDao.numByEnterprise[enterprise], total)
	}
}
