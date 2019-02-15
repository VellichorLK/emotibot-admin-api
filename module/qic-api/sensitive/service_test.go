package sensitive

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	"testing"
)

type mockDAO struct{}

var mockCategories []model.SensitiveWordCategory = []model.SensitiveWordCategory{
	model.SensitiveWordCategory{
		ID:   1234,
		Name: "n1",
	},
	model.SensitiveWordCategory{
		ID:   2345,
		Name: "n2",
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
	return 1, nil
}

func (dao *mockDAO) GetBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) ([]model.SensitiveWord, error) {
	return []model.SensitiveWord{}, nil
}

func (dao *mockDAO) CreateCateogry(category *model.SensitiveWordCategory, sqlLike model.SqlLike) (int64, error) {
	return 5, nil
}

func (dao *mockDAO) GetCategories(filter *model.SensitiveWordCategoryFilter, sqlLike model.SqlLike) ([]model.SensitiveWordCategory, error) {
	return mockCategories, nil
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

func setupSensitiveWordMock() (model.DBLike, model.SensitiveWordDao, model.SentenceDao) {
	originDBLike := dbLike
	dbLike = &test.MockDBLike{}

	mockdao := &mockDAO{}
	originDao := swDao
	swDao = mockdao

	originSentenceDao := sentenceDao
	sentenceDao = mockdao
	return originDBLike, originDao, originSentenceDao
}

func restoreSensitiveWordMock(originDBLike model.DBLike, originDao model.SensitiveWordDao, originSDao model.SentenceDao) {
	dbLike = originDBLike
	swDao = originDao
	sentenceDao = originSDao
}

func TestIsSensitive(t *testing.T) {
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
	originDBLike, originDao, originSDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao)

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
	originDBLike, originDao, originSDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao)

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
	originDBLike, originDao, originSDao := setupSensitiveWordMock()
	defer restoreSensitiveWordMock(originDBLike, originDao, originSDao)

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
