package sensitive

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"github.com/anknown/ahocorasick"
)

var dao sensitiveDao = &sensitiveDAOImpl{}

func IsSensitive(content string) ([]string, error) {
	matched := []string{}
	words, err := dao.GetSensitiveWords()
	if err != nil {
		return matched, err
	}

	rwords := stringsToRunes(words)

	m := new(goahocorasick.Machine)
	if err = m.Build(rwords); err != nil {
		return matched, err
	}

	terms := m.MultiPatternSearch([]rune(content), false)
	for _, t := range terms {
		matched = append(matched, string(t.Word))
	}

	return matched, nil
}

func stringsToRunes(ss []string) [][]rune {
	words := make([][]rune, len(ss), len(ss))
	for idx, s := range ss {
		word := []rune(s)
		words[idx] = word
	}
	return words
}

func CreateSensitiveWord(name, enterprise string, score int, customerException, staffException []string) (uid string, err error) {
	uid, err = general.UUID()
	if err != nil {
		return
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	word := &model.SensitiveWord{
		Name:       name,
		Enterprise: enterprise,
		Score:      score,
	}

	var deleted int8
	sq := &model.SentenceQuery{
		UUID:       customerException,
		IsDelete:   &deleted,
		Enterprise: &enterprise,
		Limit:      100,
	}

	customerExceptionSentences, err := sentenceDao.GetSentences(tx, sq)
	if err != nil {
		return
	}
	word.CustomerException = model.ToSimpleSentences(customerExceptionSentences)

	sq.UUID = staffException
	staffExceptionSentences, err := sentenceDao.GetSentences(tx, sq)
	if err != nil {
		return
	}
	word.StaffException = model.ToSimpleSentences(staffExceptionSentences)

	rowID, err := swDao.Create(word, tx)
	if err != nil {
		return
	}

	word.ID = rowID
	word.UUID = uid
	err = dbLike.Commit(tx)
	return
}

func CreateSensitiveWordCategory(name, enterprise string) (int64, error) {
	sqlConn := dbLike.Conn()

	category := &model.SensitiveWordCategory{
		Name:       name,
		Enterprise: enterprise,
	}
	return swDao.CreateCateogry(category, sqlConn)
}

func GetCategories(enterprise string) (categories []model.SensitiveWordCategory, err error) {
	sqlConn := dbLike.Conn()

	filter := &model.SensitiveWordCategoryFilter{
		ID:         []int64{},
		Enterprise: &enterprise,
	}

	return swDao.GetCategories(filter, sqlConn)
}
