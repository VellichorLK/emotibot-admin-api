package sensitive

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"fmt"
	"github.com/anknown/ahocorasick"
)

var dao sensitiveDao = &sensitiveDAOImpl{}
var (
	ErrZeroAffectedRows = fmt.Errorf("No rows are affected")
)

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

	word.UUID = uid
	rowID, err := swDao.Create(word, tx)
	if err != nil {
		return
	}

	word.ID = rowID
	err = dbLike.Commit(tx)
	return
}

func GetSensitiveWords(filter *model.SensitiveWordFilter) (total int64, words []model.SensitiveWord, err error) {
	sqlConn := dbLike.Conn()

	total, err = swDao.CountBy(filter, sqlConn)
	if err != nil {
		return
	}

	if total == 0 {
		words = []model.SensitiveWord{}
		return
	}

	words, err = swDao.GetBy(filter, sqlConn)
	return
}

func GetSensitiveWordInDetail(wUUID string, enterprise string) (word *model.SensitiveWord, err error) {
	sqlConn := dbLike.Conn()

	var deleted int8
	filter := &model.SensitiveWordFilter{
		UUID:       []string{wUUID},
		Enterprise: &enterprise,
		Deleted:    &deleted,
	}

	words, err := swDao.GetBy(filter, sqlConn)
	if err != nil {
		return
	}

	if len(words) == 0 {
		return
	}

	word = &words[0]

	rels, err := swDao.GetRel(word.ID, sqlConn)
	if err != nil {
		return
	}

	if staffException, ok := rels[model.StaffExceptionType]; ok {
		query := &model.SentenceQuery{
			ID:         staffException,
			Enterprise: &enterprise,
		}

		sens, err := sentenceDao.GetSentences(sqlConn, query)
		if err != nil {
			return nil, err
		}

		word.StaffException = model.ToSimpleSentences(sens)
	}

	if customerException, ok := rels[model.CustomerExceptionType]; ok {
		query := &model.SentenceQuery{
			ID:         customerException,
			Enterprise: &enterprise,
		}

		sens, err := sentenceDao.GetSentences(sqlConn, query)
		if err != nil {
			return nil, err
		}

		word.CustomerException = model.ToSimpleSentences(sens)
	}
	return
}

func CreateSensitiveWordCategory(name, enterprise string) (int64, error) {
	sqlConn := dbLike.Conn()

	category := &model.CategoryRequest{
		Name:       name,
		Type:       model.SwCategoryType,
		Enterprise: enterprise,
	}
	return categoryDao.InsertCategory(sqlConn, category)
}

func GetCategories(enterprise string) (categories []*model.CategortInfo, err error) {
	sqlConn := dbLike.Conn()

	ctype := model.SwCategoryType
	filter := &model.CategoryQuery{
		ID:         []uint64{},
		Enterprise: &enterprise,
		Type:       &ctype,
	}

	return categoryDao.GetCategories(sqlConn, filter)
}

func UpdateCategory(id int64, category *model.CategoryRequest) (err error) {
	sqlConn := dbLike.Conn()
	return categoryDao.UpdateCategory(sqlConn, uint64(id), category)
}

func DeleteCategory(id int64, enterprise string) (affectRows int64, err error) {
	sqlConn := dbLike.Conn()
	filter := &model.CategoryQuery{
		ID:         []uint64{uint64(id)},
		Enterprise: &enterprise,
	}

	affectRows, err = categoryDao.SoftDeleteCategory(sqlConn, filter)
	if err != nil {
		return
	}

	if affectRows == 0 {
		err = ErrZeroAffectedRows
	}
	return
}
