package sensitive

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	"testing"
)

type mockDAO struct{}

var mockCategories []*model.CategortInfo = []*model.CategortInfo{
	&model.CategortInfo{
		ID:   1234,
		Name: "n1",
	},
	&model.CategortInfo{
		ID:   2345,
		Name: "n2",
	},
}

var mockWords []model.SensitiveWord = []model.SensitiveWord{
	model.SensitiveWord{
		ID:         1234,
		UUID:       "1234",
		Name:       "n1",
		Enterprise: "abcd",
		CategoryID: 5,
	},
	model.SensitiveWord{
		ID:         5678,
		UUID:       "5678",
		Name:       "n2",
		Enterprise: "cddc",
		CategoryID: 5,
	},
}

func (dao *mockDAO) GetSensitiveWords() ([]string, error) {
	return []string{
		"收益",
	}, nil
}

func (dao *mockDAO) Create(word *model.SensitiveWord, sqlLike model.SqlLike) (int64, error) {
	return 1, nil

}

func (dao *mockDAO) CountBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) (int64, error) {
	return int64(len(mockWords)), nil
}

func (dao *mockDAO) GetBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) ([]model.SensitiveWord, error) {
	return mockWords, nil
}

func (dao *mockDAO) GetRel(id int64, sqlLike model.SqlLike) (map[int8][]uint64, error) {
	return map[int8][]uint64{}, nil
}

func (dao *mockDAO) Delete(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
}

func (dao *mockDAO) Move(filter *model.SensitiveWordFilter, categoryID int64, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
}

func (dao *mockDAO) Names(sqlLike model.SqlLike, forced bool) ([]string, error) {
	return []string{
		"收益",
	}, nil
}

func (dao *mockDAO) GetCategories(conn model.SqlLike, q *model.CategoryQuery) ([]*model.CategortInfo, error) {
	return mockCategories, nil
}
func (dao *mockDAO) InsertCategory(conn model.SqlLike, s *model.CategoryRequest) (int64, error) {
	return 5, nil
}
func (dao *mockDAO) SoftDeleteCategory(conn model.SqlLike, q *model.CategoryQuery) (int64, error) {
	return 1, nil
}
func (dao *mockDAO) CountCategory(conn model.SqlLike, q *model.CategoryQuery) (uint64, error) {
	return uint64(len(mockCategories)), nil
}

func (dao *mockDAO) UpdateCategory(conn model.SqlLike, id uint64, s *model.CategoryRequest) error {
	return nil
}

func (dao *mockDAO) GetSentences(tx model.SqlLike, q *model.SentenceQuery) ([]*model.Sentence, error) {
	return []*model.Sentence{}, nil
}

func (dao *mockDAO) InsertSentence(tx model.SqlLike, s *model.Sentence) (int64, error) {
	return 1, nil
}

func (dao *mockDAO) SoftDeleteSentence(tx model.SqlLike, q *model.SentenceQuery) (int64, error) {
	return 1, nil
}

func (dao *mockDAO) CountSentences(tx model.SqlLike, q *model.SentenceQuery) (uint64, error) {
	return 1, nil
}

func (dao *mockDAO) InsertSenTagRelation(tx model.SqlLike, s *model.Sentence) error {
	return nil
}

func (dao *mockDAO) GetRelSentenceIDByTagIDs(tx model.SqlLike, tagIDs []uint64) (map[uint64][]uint64, error) {
	return map[uint64][]uint64{}, nil
}

func (dao *mockDAO) MoveCategories(x model.SqlLike, q *model.SentenceQuery, category uint64) (int64, error) {
	return 1, nil
}

func (dao *mockDAO) InsertSentences(sqlLike model.SqlLike, sentences []model.Sentence) error {
	return nil
}

var mockdao sensitiveDao = &mockDAO{}

func setupSensitiveWordMock() (model.DBLike, model.SensitiveWordDao, model.SentenceDao, model.CategoryDao) {
	originDBLike := dbLike
	dbLike = &test.MockDBLike{}

	mockdao := &mockDAO{}
	originDao := swDao
	swDao = mockdao

	originCateDao := categoryDao
	categoryDao = mockdao

	originSentenceDao := sentenceDao
	sentenceDao = mockdao

	return originDBLike, originDao, originSentenceDao, originCateDao
}

func restoreSensitiveWordMock(originDBLike model.DBLike, originDao model.SensitiveWordDao, originSDao model.SentenceDao, originCateDao model.CategoryDao) {
	dbLike = originDBLike
	swDao = originDao
	sentenceDao = originSDao
	categoryDao = originCateDao
}

func TestIsSensitive(t *testing.T) {
	originDBLike, originDao, originSDao, originCateDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao, originCateDao)

	sen1 := "收益"
	sen2 := "一個安全的句子"
	sen3 := "要不要理财型保险"

	sen1Result, _ := IsSensitive(sen1)
	sen2Result, _ := IsSensitive(sen2)
	sen3Result, _ := IsSensitive(sen3)

	if len(sen1Result) == 0 || len(sen2Result) > 0 || len(sen3Result) > 0 {
		t.Error("check sensitive words fail")
	}
}

func TestStringsToRunes(t *testing.T) {
	ss, _ := mockdao.GetSensitiveWords()
	words := stringsToRunes(ss)

	if len(words) != len(ss) {
		t.Error("tranforms strings to runes failed")
	}
}

func TestCreateSensitiveWord(t *testing.T) {
	originDBLike, originDao, originSDao, originCateDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao, originCateDao)

	word := &model.SensitiveWord{
		Score:      5,
		Enterprise: "abcd",
	}

	uid, err := CreateSensitiveWord(word.Name, word.Enterprise, word.Score, []string{}, []string{})
	if err != nil {
		t.Errorf("error when create sensitive word, err: %s", err.Error())
		return
	}

	if uid == "" {
		t.Errorf("expect uid not be empty string, but got empty")
	}
}

func TestCreateSensitiveWordCategory(t *testing.T) {
	originDBLike, originDao, originSDao, originCateDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao, originCateDao)

	name := "test again"
	enterprise := "1234"
	id, err := CreateSensitiveWordCategory(name, enterprise)
	if err != nil {
		t.Errorf("error when create sensitive word category, err: %s", err.Error())
		return
	}

	if id != 5 {
		t.Errorf("expect 5 but got: %d", id)
		return
	}
}

func TestGetCategories(t *testing.T) {
	originDBLike, originDao, originSDao, originCateDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao, originCateDao)

	categories, err := GetCategories("test")
	if err != nil {
		t.Errorf("error when test get categories, err: %s", err.Error())
		return
	}

	for idx, cate := range categories {
		targetCate := mockCategories[idx]
		if targetCate.ID != cate.ID {
			t.Errorf("expect %d but got: %d", targetCate.ID, cate.ID)
			return
		}

		if targetCate.Name != cate.Name {
			t.Errorf("expect %s but got %s", targetCate.Name, cate.Name)
			return
		}
	}
}

func TestUpdateCategory(t *testing.T) {
	originDBLike, originDao, originSDao, originCateDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao, originCateDao)

	category := &model.CategoryRequest{
		Name: "55588",
	}

	err := UpdateCategory(55, category)
	if err != nil {
		t.Errorf("error when test update category, err: %s", err.Error())
		return
	}
}

func TestDeleteCategory(t *testing.T) {
	originDBLike, originDao, originSDao, originCateDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao, originCateDao)

	affectedRows, err := DeleteCategory(55688, "55688")
	if err != nil {
		t.Errorf("error when test delete category, err: %s", err.Error())
		return
	}

	if affectedRows != 1 {
		t.Errorf("expect affectrows: %d but got %d", 1, affectedRows)
	}
}
