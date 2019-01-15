package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var categoryDao model.CategoryDao = &model.CategorySQLDao{}

func GetCategories(query *model.CategoryQuery) (categories []*model.CategortInfo, err error) {
	sqlConn := dbLike.Conn()
	return categoryDao.GetCategories(sqlConn, query)
}

func CreateCategory(category *model.CategortInfo) (id int64, err error) {
	sqlConn := dbLike.Conn()
	c := &model.CategoryRequest{
		Name:       category.Name,
		Enterprise: category.Enterprise,
	}
	return categoryDao.InsertCategory(sqlConn, c)
}

func GetCategorySentences(query *model.SentenceQuery) (total uint64, sentences []*model.Sentence, err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}

	total, err = sentenceDao.CountSentences(tx, query)
	if err != nil {
		return
	}

	sentences, err = sentenceDao.GetSentences(tx, query)
	return
}

func UpdateCategory(id uint64, category *model.CategoryRequest) (err error) {
	sqlConn := dbLike.Conn()
	return categoryDao.UpdateCategory(sqlConn, id, category)
}
