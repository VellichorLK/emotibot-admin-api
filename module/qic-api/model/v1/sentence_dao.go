package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

//SentenceDao defines the db access function
type SentenceDao interface {
	GetSentences(tx SqlLike, q *SentenceQuery) ([]*Sentence, error)
	InsertSentence(tx SqlLike, s *Sentence) (int64, error)
	SoftDeleteSentence(tx SqlLike, q *SentenceQuery) (int64, error)
	CountSentences(tx SqlLike, q *SentenceQuery) (uint64, error)
	InsertSenTagRelation(tx SqlLike, s *Sentence) error
	GetRelSentenceIDByTagIDs(tx SqlLike, tagIDs []uint64) (map[uint64][]uint64, error)
	MoveCategories(x SqlLike, q *SentenceQuery, category uint64) (int64, error)
	InsertSentences(SqlLike, []Sentence) error
}

//error message
var (
	ErrNeedCondition   = errors.New("Must has query condition")
	ErrNeedTransaction = errors.New("Must use transaction")
	ErrNeedRelation    = errors.New("Must has relation structure")
)

// SimpleSentence only contains sentence id & sentence name
type SimpleSentence struct {
	ID         uint64 `json:"-"`
	UUID       string `json:"sentence_id"`
	Name       string `json:"sentence_name"`
	CategoryID uint64 `json:"category_id,string"`
}

//SentenceQuery uses as query structure of sentence
type SentenceQuery struct {
	UUID       []string
	ID         []uint64
	Enterprise *string
	Page       int
	Limit      int
	IsDelete   *int8
	CategoryID *uint64
}

//SentenceNewRecord is used to create a new sentence
type SentenceNewRecord struct {
	TagID []uint64 `json:"tags"`
	Name  string   `json:"sentence_name"`
}

//SentenceSQLDao implements the sql acess
type SentenceSQLDao struct {
	conn *sql.DB
}

//Sentence is sentence data in db
type Sentence struct {
	ID         uint64
	CategoryID uint64
	IsDelete   int8
	Name       string
	Enterprise string
	UUID       string
	CreateTime int64
	UpdateTime int64
	TagIDs     []uint64
}

func ToSimpleSentences(ss []*Sentence) []SimpleSentence {
	simpleSentences := make([]SimpleSentence, len(ss))
	for idx, s := range ss {
		simpleSentences[idx] = SimpleSentence{
			Name:       s.Name,
			ID:         s.ID,
			UUID:       s.UUID,
			CategoryID: s.CategoryID,
		}
	}
	return simpleSentences
}

//NewSentenceSQLDao generates the structure of SentenceSQLDao
func NewSentenceSQLDao(conn *sql.DB) *SentenceSQLDao {
	return &SentenceSQLDao{
		conn: conn,
	}
}

func genrateExecutor(conn *sql.DB, tx SqlLike) (SqlLike, error) {
	var q SqlLike
	if tx != nil {
		q = tx
	} else if conn != nil {
		q = conn
	} else {
		return nil, util.ErrDBNotInit
	}
	return q, nil
}

func (q *SentenceQuery) whereSQL() (string, []interface{}) {
	numOfUUID := len(q.UUID)
	params := make([]interface{}, 0, numOfUUID+1)
	conditions := []string{}

	if numOfUUID > 0 {
		condition := fldUUID + " IN (?" + strings.Repeat(",?", numOfUUID-1) + ")"
		conditions = append(conditions, condition)
		for i := 0; i < numOfUUID; i++ {
			params = append(params, q.UUID[i])
		}
	}

	numOfID := len(q.ID)
	if numOfID > 0 {
		condition := fldID + " IN (?" + strings.Repeat(",?", numOfID-1) + ")"
		conditions = append(conditions, condition)
		for i := 0; i < numOfID; i++ {
			params = append(params, q.ID[i])
		}
	}

	if q.Enterprise != nil {
		condition := fldEnterprise + " = ?"
		conditions = append(conditions, condition)
		params = append(params, *q.Enterprise)
	}

	if q.IsDelete != nil {
		condition := fldIsDelete + "=?"
		conditions = append(conditions, condition)
		params = append(params, *q.IsDelete)
	}

	if q.CategoryID != nil {
		condition := fldCategoryID + "=?"
		conditions = append(conditions, condition)
		params = append(params, *q.CategoryID)
	}

	var whereSQL string
	if len(conditions) > 0 {
		whereSQL = "WHERE " + strings.Join(conditions, " AND ")
	}

	whereSQL += " ORDER BY " + fldID + " DESC "

	if q.Page > 0 && q.Limit > 0 {
		whereSQL = fmt.Sprintf("%s LIMIT %d OFFSET %d", whereSQL, q.Limit, (q.Page-1)*q.Limit)
	}

	return whereSQL, params
}

//GetSentences gets the sentences based on query condition
func (d *SentenceSQLDao) GetSentences(tx SqlLike, sq *SentenceQuery) ([]*Sentence, error) {
	q, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return nil, err
	}

	var condition string
	var params []interface{}
	if sq != nil {
		condition, params = sq.whereSQL()
	}

	queryStr := fmt.Sprintf("SELECT a.%s,a.%s,a.%s,a.%s,a.%s,a.%s,a.%s,a.%s,b.%s FROM (SELECT * FROM %s %s) as a LEFT JOIN %s as b on a.%s=b.%s",
		fldID, fldIsDelete, fldName, fldEnterprise, fldUUID, fldCreateTime, fldUpdateTime, fldCategoryID, fldRelTagID,
		tblSentence, condition,
		tblRelSenTag, fldID, fldRelSenID)

	rows, err := q.Query(queryStr, params...)
	if err != nil {
		logger.Error.Printf("Query: %s, Params:%+v, failed. %s\n", queryStr, params, err)
		return nil, err
	}
	defer rows.Close()

	sentences := make([]*Sentence, 0, 10)
	existMap := make(map[uint64]*Sentence)
	for rows.Next() {
		var s Sentence
		var tagID sql.NullInt64
		err = rows.Scan(&s.ID, &s.IsDelete, &s.Name, &s.Enterprise, &s.UUID, &s.CreateTime, &s.UpdateTime, &s.CategoryID, &tagID)
		if err != nil {
			logger.Error.Printf("Scan error. %s\n", err)
			return nil, err
		}

		if sentence, ok := existMap[s.ID]; ok {
			if tagID.Valid {
				sentence.TagIDs = append(sentence.TagIDs, uint64(tagID.Int64))
			}
		} else {
			s.TagIDs = make([]uint64, 0, 2)
			if tagID.Valid {
				s.TagIDs = append(s.TagIDs, uint64(tagID.Int64))
			}
			sentences = append(sentences, &s)
			existMap[s.ID] = &s
		}
	}
	return sentences, nil
}

func getSentenceInsertSQL(sentences []Sentence) (insertStr string, values []interface{}) {
	values = []interface{}{}
	if len(sentences) == 0 {
		return
	}

	fields := []string{
		fldIsDelete, fldName, fldEnterprise, fldUUID, fldCreateTime, fldUpdateTime, fldCategoryID,
	}
	fieldStr := strings.Join(fields, ", ")

	variableStr := fmt.Sprintf("(?%s)", strings.Repeat(", ?", len(fields)-1))
	valueStr := ""
	for _, s := range sentences {
		values = append(
			values,
			s.IsDelete, s.Name, s.Enterprise, s.UUID, s.CreateTime, s.UpdateTime, s.CategoryID,
		)
		valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tblSentence,
		fieldStr,
		valueStr,
	)
	return
}

//InsertSentence inserts a new sentence and the relation between sentence and tag
//if you insert both, must use transaction
func (d *SentenceSQLDao) InsertSentence(tx SqlLike, s *Sentence) (int64, error) {
	if s == nil {
		return 0, nil
	}

	numOfTags := len(s.TagIDs)
	if numOfTags > 0 && tx == nil {
		return 0, ErrNeedTransaction
	}

	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return 0, err
	}

	//insert into Sentence table
	insertSenSQL, values := getSentenceInsertSQL([]Sentence{*s})

	res, err := exe.Exec(insertSenSQL, values...)
	if err != nil {
		logger.Error.Printf("insert sentence %s failed %s\n", insertSenSQL, err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.Error.Printf("Acquire last insert id failed. %s\n", err)
		return 0, err
	}

	return id, err
}

//InsertSenTagRelation inserts the Relation_Sentence_Tag
func (d *SentenceSQLDao) InsertSenTagRelation(tx SqlLike, s *Sentence) error {
	if s == nil || len(s.TagIDs) == 0 || s.ID == 0 {
		return ErrNeedRelation
	}

	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return err
	}

	numOfTags := len(s.TagIDs)
	insertRelSQL := fmt.Sprintf("INSERT INTO %s (%s,%s) VALUES ", tbleRelSentenceTag, fldRelSenID, fldRelTagID)

	bulk := fmt.Sprintf("(%d,?)", s.ID)
	insertRelSQL = fmt.Sprintf("%s %s%s", insertRelSQL, bulk, strings.Repeat(","+bulk, numOfTags-1))
	params := make([]interface{}, 0, numOfTags)
	for i := 0; i < numOfTags; i++ {
		params = append(params, s.TagIDs[i])
	}
	_, err = exe.Exec(insertRelSQL, params...)
	if err != nil {
		logger.Error.Printf("insert (%s)(%+v) relation sentence tag failed. %s\n", insertRelSQL, params, err)
		return err
	}
	return nil
}

//SoftDeleteSentence sets the field is_delete to 1
func (d *SentenceSQLDao) SoftDeleteSentence(tx SqlLike, q *SentenceQuery) (int64, error) {
	if q == nil || q.Enterprise == nil || len(q.UUID) == 0 {
		return 0, ErrNeedCondition
	}
	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return 0, err
	}
	condition, params := q.whereSQL()
	deleteSQL := fmt.Sprintf("Update %s SET %s=1,%s=%d %s", tblSentence, fldIsDelete, fldUpdateTime,
		time.Now().Unix(), condition)

	res, err := exe.Exec(deleteSQL, params...)
	if err != nil {
		logger.Error.Printf("delete failed. %s\n", err)
		return 0, err
	}
	return res.RowsAffected()
}

//CountSentences counts number of records
func (d *SentenceSQLDao) CountSentences(tx SqlLike, q *SentenceQuery) (uint64, error) {
	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return 0, err
	}
	condition, params := q.whereSQL()
	querySQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tblSentence, condition)
	var count uint64
	err = exe.QueryRow(querySQL, params...).Scan(&count)
	if err != nil {
		logger.Error.Printf("Query row (%s) failed. %s\n", querySQL, err)
		return 0, err
	}
	return count, nil
}

//GetRelSentenceIDByTagIDs gets the sentence id which is related to tag id
func (d *SentenceSQLDao) GetRelSentenceIDByTagIDs(tx SqlLike, tagIDs []uint64) (map[uint64][]uint64, error) {
	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return nil, err
	}

	numOfTags := len(tagIDs)
	if numOfTags == 0 {
		return nil, nil
	}

	querySQL := fmt.Sprintf("SELECT %s,%s FROM %s WHERE %s IN (?%s)",
		fldRelSenID, fldRelTagID,
		tblRelSenTag, fldRelTagID,
		strings.Repeat(",?", numOfTags-1))

	tagIDsInterface := make([]interface{}, 0, numOfTags)
	for _, id := range tagIDs {
		tagIDsInterface = append(tagIDsInterface, id)
	}

	rows, err := exe.Query(querySQL, tagIDsInterface...)
	if err != nil {
		logger.Error.Printf("get sentence to tag relation failed. %s\n", err)
		return nil, err
	}
	defer rows.Close()

	tagToSentenceMap := make(map[uint64][]uint64)
	var sID, tID uint64
	var sentenceIDs []uint64
	var ok bool
	for rows.Next() {
		err = rows.Scan(&sID, &tID)
		if err != nil {
			logger.Error.Printf("Scan error. %s\n", err)
			return nil, err
		}
		if sentenceIDs, ok = tagToSentenceMap[tID]; !ok {
			sentenceIDs = make([]uint64, 0, 1)
		}
		sentenceIDs = append(sentenceIDs, sID)
		tagToSentenceMap[tID] = sentenceIDs
	}

	return tagToSentenceMap, nil
}

//MoveCategories updates the category_id for given sentences
func (d *SentenceSQLDao) MoveCategories(tx SqlLike, q *SentenceQuery, category uint64) (int64, error) {
	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return 0, err
	}

	if q == nil {
		return 0, ErrNeedCondition
	}

	condition, params := q.whereSQL()
	updateSQL := fmt.Sprintf("UPDATE %s SET %s=? %s",
		tblSentence,
		fldCategoryID, condition)

	params = append([]interface{}{category}, params...)
	res, err := exe.Exec(updateSQL, params...)
	if err != nil {
		logger.Error.Printf("update the category failed. %s\n %s %d\n", err, updateSQL, category)
		return 0, err
	}
	return res.RowsAffected()
}

func (d *SentenceSQLDao) InsertSentences(sqlLike SqlLike, sentences []Sentence) (err error) {
	if len(sentences) == 0 {
		return
	}

	insertStr, values := getSentenceInsertSQL(sentences)
	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("insert sentences failed. sql: %s\n", insertStr)
		logger.Error.Printf("values: %+v\n", values)
		return
	}

	varaibleStr := "(?, ?)"
	valueStr := ""
	values = []interface{}{}
	sUUID := []string{}
	for _, s := range sentences {
		for _ = range s.TagIDs {
			valueStr = fmt.Sprintf("%s%s,", valueStr, varaibleStr)
		}
		sUUID = append(sUUID, s.UUID)
	}

	if valueStr == "" {
		return
	}

	var deleted int8
	sentenceQuery := &SentenceQuery{
		UUID:     sUUID,
		IsDelete: &deleted,
	}

	newSentences, err := d.GetSentences(sqlLike, sentenceQuery)
	if err != nil {
		return
	}

	sentenceMap := map[string]uint64{}
	for _, s := range newSentences {
		sentenceMap[s.UUID] = s.ID
	}

	for _, s := range sentences {
		for _, tag := range s.TagIDs {
			values = append(
				values,
				sentenceMap[s.UUID],
				tag,
			)
		}
	}

	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES %s",
		tbleRelSentenceTag,
		fldRelSenID,
		fldRelTagID,
		valueStr,
	)

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("insert sentences relations failed. sql: %s\n", insertStr)
		logger.Error.Printf("values: %+v\n", values)
	}
	return
}
