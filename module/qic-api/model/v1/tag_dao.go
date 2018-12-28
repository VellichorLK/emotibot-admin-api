package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/misc/mathutil"
)

//TagSQLDao is the DAO implemented in sql.
type TagSQLDao struct {
	db *sql.DB
}

type tagSQLTransaction struct {
	tx *sql.Tx
}

// NewTagSQLDao create a new tag sql dao, and check the TABLE already exist or not.
// return error if table can not be found.
func NewTagSQLDao(db *sql.DB) (*TagSQLDao, error) {
	if db == nil {
		return nil, fmt.Errorf("db can not be nil")
	}
	row := db.QueryRow("SHOW TABLEs LIKE '" + tblTags + "'")
	var tblName string
	if err := row.Scan(&tblName); err != nil || tblName != tblTags {
		return nil, fmt.Errorf("expect to have table %s, but got err: %v", tblTags, err)
	}
	return &TagSQLDao{db: db}, nil
}

/*
func Transaction(dao *tagSQLDao) (tagDao TagDao, commitcb func() error, err error) {
	tx, err := dao.db.Begin()
	if err != nil {
		return nil, func() error { return fmt.Errorf("no-op error") }, err
	}
	return &tagSQLTransaction{tx: tx}, tx.Commit, nil
}*/

//Tag is the basic unit of the training rule. which also will be used in cu training model.
type Tag struct {
	// ID is the inner id for the tag. which is changed after each modify.
	ID uint64
	// UUID is the presentation id for the tag, which will not changed after created
	UUID string
	// Enterprise is the id of enterprise the tag belong to.
	Enterprise string
	// Name should be unique in the tags of the same enterprise.
	Name string
	// Typ is the status
	Typ              int8
	PositiveSentence string
	NegativeSentence string
	// IsDelete is the soft delete flag.
	IsDeleted  bool
	CreateTime int64
	UpdateTime int64
}

// TagQuery is the query against tag datastore, all fields is the condition which should be nilable.
// any nil field will be ignored. And all conditions will be join by AND logic.
type TagQuery struct {
	ID               []uint64
	UUID             []string
	Enterprise       *string
	Name             *string
	IgnoreSoftDelete bool
	Paging           *Pagination
}

// Limit and Page is the conditions for querying paging.
type Pagination struct {
	Limit int
	Page  int
}

func (t *TagQuery) whereSQL() (string, []interface{}) {
	var conditions []string
	bindedData := make([]interface{}, 0)
	if len(t.ID) > 0 {
		conditions = append(conditions, fldTagID+" IN (?"+strings.Repeat(", ?", len(t.ID)-1)+")")
		for _, id := range t.ID {
			bindedData = append(bindedData, id)
		}
	}
	if len(t.UUID) > 0 {
		conditions = append(conditions, fldTagUUID+" IN (?"+strings.Repeat(", ?", len(t.UUID)-1)+")")
		for _, uuid := range t.UUID {
			bindedData = append(bindedData, uuid)
		}
	}
	if t.Name != nil {
		conditions = append(conditions, fldTagName+" LIKE ?")
		bindedData = append(bindedData, t.Name)
	}
	if !t.IgnoreSoftDelete {
		conditions = append(conditions, fldTagIsDeleted+" = ?")
		bindedData = append(bindedData, t.IgnoreSoftDelete)
	}

	var rawsql string
	if len(conditions) > 0 {
		rawsql = "WHERE " + strings.Join(conditions, " AND ")
	}
	return rawsql, bindedData
}

func (p *Pagination) offsetSQL() string {
	limit := mathutil.MaxInt(p.Limit, 0)
	page := mathutil.MaxInt(p.Page-1, 1)
	offset := limit * page
	return fmt.Sprintf(" LIMIT %d, %d", offset, limit)
}

func (t *TagSQLDao) Begin() (*sql.Tx, error) {
	return t.db.Begin()
}

type queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

var tagSelectColumns = []string{fldTagID, fldTagUUID, fldTagIsDeleted, fldTagName,
	fldTagType, fldTagPosSen, fldTagNegSen, fldTagCreateTime, fldTagUpdateTime, fldTagEnterprise}

// Tags fetch the tag resource from db or tx.
// query determine condition and how many it should fetch.
func (t *TagSQLDao) Tags(tx *sql.Tx, query TagQuery) ([]Tag, error) {
	var q SqlLike
	if tx != nil {
		q = tx
	} else {
		q = t.db
	}
	var (
		wheresql, pagingsql string
		data                []interface{}
	)
	wheresql, data = query.whereSQL()
	if query.Paging != nil {
		pagingsql = query.Paging.offsetSQL()
	}
	rawsql := "SELECT `" + strings.Join(tagSelectColumns, "`, `") + "` FROM `" + tblTags + "` " +
		wheresql + " " + pagingsql
	rows, err := q.Query(rawsql, data...)
	if err != nil {
		logger.Error.Println("raw error sql of ", rawsql, "; data: ", data)
		return nil, fmt.Errorf("Query error: %v", err)
	}
	defer rows.Close()
	tags := make([]Tag, 0)
	for rows.Next() {
		var (
			tag    Tag
			negSen sql.NullString
			posSen sql.NullString
		)
		rows.Scan(&tag.ID, &tag.UUID, &tag.IsDeleted, &tag.Name,
			&tag.Typ, &posSen, &negSen, &tag.CreateTime, &tag.UpdateTime, &tag.Enterprise)
		if !negSen.Valid {
			tag.NegativeSentence = "[]"
		}
		if !posSen.Valid {
			tag.PositiveSentence = "[]"
		}
		tag.PositiveSentence = posSen.String
		tag.NegativeSentence = negSen.String
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}
	return tags, nil
}

// NewTags Insert the tag struct and return the tag inserted with latest id.
// ErrAutoIDDisabled is returned If LastInsertId is not guaranteed.
func (t *TagSQLDao) NewTags(tx *sql.Tx, tags []Tag) ([]Tag, error) {
	var s SqlLike
	if tx != nil {
		s = tx
	} else {
		s = t.db
	}
	tagInsertColumns := []string{fldTagUUID, fldTagEnterprise, fldTagName, fldTagType,
		fldTagPosSen, fldTagNegSen}
	rawsql := "INSERT INTO `" + tblTags + "`(`" + strings.Join(tagInsertColumns, "` , `") + "`) VALUE (?" + strings.Repeat(", ?", len(tagInsertColumns)-1) + ")"
	stmt, err := s.Prepare(rawsql)
	if err != nil {
		logger.Error.Println("error rawsql: ", rawsql)
		return nil, fmt.Errorf("prepare sql failed, %v", err)
	}
	defer stmt.Close()
	results := make([]Tag, 0, len(tags))
	var isReliable = true
	for _, t := range tags {
		result, err := stmt.Exec(t.UUID, t.Enterprise, t.Name, t.Typ,
			t.PositiveSentence, t.NegativeSentence)
		if err != nil {
			return nil, fmt.Errorf("insert sql failed, %v", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			isReliable = false
		} else {
			t.ID = uint64(id)
		}
		results = append(results, t)
	}
	if !isReliable {
		return results, ErrAutoIDDisabled
	}
	return results, nil
}

// DeleteTags soft delete the tags from db by the query indicated.
// soft deleted tags will still stayed in the db, but isDeleted will be flaged as true.
// the true affected rows will be returned.
// notice: query's paging struct still can affected how many it will deleted.
func (t *TagSQLDao) DeleteTags(tx *sql.Tx, query TagQuery) (int64, error) {
	var q SqlLike
	if tx != nil {
		q = tx
	} else {
		q = t.db
	}
	var (
		wheresql, pagingsql string
		data                []interface{}
	)
	wheresql, data = query.whereSQL()
	if query.Paging != nil {
		pagingsql = query.Paging.offsetSQL()
	}
	rawsql := "UPDATE `" + tblTags + "` SET `" + fldTagIsDeleted + "`=?, `" + fldTagUpdateTime + "`=? " +
		wheresql + " " + pagingsql
	utctimestamp := time.Now().Unix()
	data = append([]interface{}{true, utctimestamp}, data...)
	result, err := q.Exec(rawsql, data...)
	if err != nil {
		logger.Error.Println("error sql: " + rawsql)
		return 0, fmt.Errorf("update failed, %v", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, ErrAutoIDDisabled
	}
	return count, nil
}

func (t *TagSQLDao) CountTags(tx *sql.Tx, query TagQuery) (uint, error) {
	var q SqlLike
	if tx != nil {
		q = tx
	} else {
		q = t.db
	}
	var (
		wheresql, pagingsql string
		data                []interface{}
	)
	wheresql, data = query.whereSQL()
	if query.Paging != nil {
		pagingsql = query.Paging.offsetSQL()
	}
	rawsql := "SELECT count(*) FROM `" + tblTags + "` " +
		wheresql + " " + pagingsql

	var count uint
	err := q.QueryRow(rawsql, data...).Scan(&count)
	if err != nil {
		logger.Error.Println("raw error sql of ", rawsql)
		return 0, fmt.Errorf("Query error: %v", err)
	}
	return count, nil
}
