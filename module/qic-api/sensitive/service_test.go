package sensitive

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	"testing"
)

type mockDAO struct{}

func (dao *mockDAO) GetSensitiveWords() ([]string, error) {
	return []string{
		"收益",
	}, nil
}

func (dao *mockDAO) Create(word *model.SensitiveWord, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
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
