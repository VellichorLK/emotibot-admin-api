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
	sqlConn := dbLike.Conn()
	words, err := swDao.Names(sqlConn, false)
	if err != nil {
		return matched, err
	}

	if len(words) == 0 {
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
	words := make([][]rune, len(ss))
	for idx, s := range ss {
		// ignore empty string
		// ignroe empty string will cause Index out of error in goahocorasick.Machine Build
		if s == "" {
			continue
		}
		word := []rune(s)
		words[idx] = word
	}
	return words
}

// CreateSensitiveWord create a uuid and create a new sensitive word
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

	customerExceptionSentences, staffExceptionSentences, err := getWordExceptionSentences(customerException, staffException, enterprise, tx)
	if err != nil {
		return
	}
	word.CustomerException = customerExceptionSentences
	word.StaffException = staffExceptionSentences

	word.UUID = uid
	rowID, err := swDao.Create(word, tx)
	if err != nil {
		return
	}

	word.ID = rowID
	err = dbLike.Commit(tx)
	return
}

// getWordExceptionSentences takes sentence uuid string slice as inputs
// and output simple sentence slice of customer exception sentences and staff exception sentences
func getWordExceptionSentences(customerSentences, staffSentences []string, enterprise string, sqlLike model.SqlLike) ([]model.SimpleSentence, []model.SimpleSentence, error) {
	customerException := []model.SimpleSentence{}
	staffException := []model.SimpleSentence{}

	var deleted int8
	sq := &model.SentenceQuery{
		UUID:       customerSentences,
		IsDelete:   &deleted,
		Enterprise: &enterprise,
		Limit:      100,
	}

	customerExceptionSentences, err := sentenceDao.GetSentences(sqlLike, sq)
	if err != nil {
		return customerException, staffException, err
	}
	customerException = model.ToSimpleSentences(customerExceptionSentences)

	sq.UUID = staffSentences
	staffExceptionSentences, err := sentenceDao.GetSentences(sqlLike, sq)
	if err != nil {
		return customerException, staffException, err
	}
	staffException = model.ToSimpleSentences(staffExceptionSentences)

	return customerException, staffException, nil
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

func UpdateSensitiveWord(word *model.SensitiveWord) (err error) {
	if word == nil {
		return
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}

	err = deleteSensitiveWord(word.UUID, word.Enterprise, tx)
	if err != nil {
		return
	}

	customerException := toStringSlice(word.CustomerException)
	staffException := toStringSlice(word.StaffException)

	customerExceptionSentences, staffExceptionSentences, err := getWordExceptionSentences(customerException, staffException, word.Enterprise, tx)
	if err != nil {
		return
	}

	word.CustomerException = customerExceptionSentences
	word.StaffException = staffExceptionSentences

	_, err = swDao.Create(word, tx)
	if err != nil {
		return
	}
	err = dbLike.Commit(tx)
	return
}

func toStringSlice(simpleSens []model.SimpleSentence) []string {
	UUID := make([]string, len(simpleSens))
	for idx, ss := range simpleSens {
		UUID[idx] = ss.UUID
	}
	return UUID
}

func DeleteSensitiveWord(uid, enterprise string) error {
	sqlConn := dbLike.Conn()

	return deleteSensitiveWord(uid, enterprise, sqlConn)
}

func deleteSensitiveWord(uid, enterprise string, sqlLike model.SqlLike) (err error) {
	var deleted int8
	filter := &model.SensitiveWordFilter{
		UUID:       []string{uid},
		Enterprise: &enterprise,
		Deleted:    &deleted,
	}

	affectedRows, err := swDao.Delete(filter, sqlLike)
	if err != nil {
		return
	}

	if affectedRows == 0 {
		err = ErrZeroAffectedRows
	}
	return
}

func MoveSensitiveWord(UUID []string, enterprise string, categoryID int64) (err error) {
	sqlConn := dbLike.Conn()

	filter := &model.SensitiveWordFilter{
		UUID:       UUID,
		Enterprise: &enterprise,
	}

	_, err = swDao.Move(filter, categoryID, sqlConn)
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
	var deleted int8
	filter := &model.CategoryQuery{
		ID:         []uint64{},
		Enterprise: &enterprise,
		Type:       &ctype,
		IsDelete:   &deleted,
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
