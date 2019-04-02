package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

type TagDao interface {
	Tags(tx SqlLike, query TagQuery) ([]Tag, error)
	NewTags(tx SqlLike, tags []Tag) ([]Tag, error)
	DeleteTags(tx SqlLike, query TagQuery) (int64, error)
	CountTags(tx SqlLike, query TagQuery) (uint, error)
}

//TagSQLDao is the DAO implemented in sql.
type TagSQLDao struct {
	db *sql.DB
}

type tagSQLTransaction struct {
	tx SqlLike
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
	ID               uint64 // the inner id for the tag. which is changed after each modify.
	IsDeleted        bool   // IsDelete is the soft delete flag.
	Name             string // the unique name in the tags of the same enterprise.
	Typ              int8   // status
	PositiveSentence string // json payload of positive data, cu training will need at least one element
	NegativeSentence string // json payload of negative data
	CreateTime       int64
	UpdateTime       int64
	Enterprise       string // the id of enterprise the tag belong to.
	UUID             string // the presentation id for the tag, which will not changed after created
}

// TagQuery is the query against tag datastore, all fields is the condition which should be nilable.
// any nil field will be ignored. And all conditions will be join by AND logic.
type TagQuery struct {
	ID               []uint64
	UUID             []string
	Enterprise       *string
	Name             *string
	TagType          []int8
	UpdateTimeStart  int64
	UpdateTimeEnd    int64
	IgnoreSoftDelete bool
	Paging           *Pagination
}

func (t *TagQuery) whereSQL() (string, []interface{}) {
	builder := NewWhereBuilder(andLogic, "")
	builder.In(fldTagID, uint64ToWildCard(t.ID...))
	builder.In(fldTagUUID, stringToWildCard(t.UUID...))
	if t.Enterprise != nil {
		builder.Eq(fldTagEnterprise, *t.Enterprise)
	}
	if t.Name != nil {
		builder.Like(fldTagName, *t.Name)
	}
	if t.UpdateTimeStart > 0 {
		builder.Gte(fldTagUpdateTime, t.UpdateTimeStart)
	}
	if t.UpdateTimeEnd > 0 {
		builder.Lt(fldTagUpdateTime, t.UpdateTimeEnd)
	}
	if len(t.TagType) > 0 {
		builder.In(fldTagType, int8ToWildCard(t.TagType...))
	}
	if !t.IgnoreSoftDelete {
		builder.Eq(fldTagIsDeleted, t.IgnoreSoftDelete)
	}

	return builder.ParseWithWhere()
}

type queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

var tagSelectColumns = []string{
	fldTagID, fldTagUUID, fldTagIsDeleted,
	fldTagName, fldTagType, fldTagPosSen,
	fldTagNegSen, fldTagCreateTime, fldTagUpdateTime,
	fldTagEnterprise}

// Tags fetch the tag resource from db or tx.
// query determine condition and how many it should fetch.
func (t *TagSQLDao) Tags(tx SqlLike, query TagQuery) ([]Tag, error) {
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
func (t *TagSQLDao) NewTags(tx SqlLike, tags []Tag) ([]Tag, error) {
	var s SqlLike
	if tx != nil {
		s = tx
	} else {
		s = t.db
	}
	tagInsertColumns := []string{
		fldTagUUID, fldTagEnterprise, fldTagName,
		fldTagType, fldTagPosSen, fldTagNegSen,
		fldTagCreateTime, fldTagUpdateTime,
	}
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
		result, err := stmt.Exec(
			t.UUID, t.Enterprise, t.Name,
			t.Typ, t.PositiveSentence, t.NegativeSentence,
			t.CreateTime, t.UpdateTime,
		)
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
func (t *TagSQLDao) DeleteTags(tx SqlLike, query TagQuery) (int64, error) {
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

// CountTags counts the tags by the given query.
// It will ignore the Paging in query to return a true count by the query conditions.
func (t *TagSQLDao) CountTags(tx SqlLike, query TagQuery) (uint, error) {
	var q SqlLike
	if tx != nil {
		q = tx
	} else {
		q = t.db
	}
	var (
		wheresql string
		data     []interface{}
	)
	wheresql, data = query.whereSQL()
	rawsql := "SELECT count(*) FROM `" + tblTags + "` " +
		wheresql

	var count uint
	err := q.QueryRow(rawsql, data...).Scan(&count)
	if err != nil {
		logger.Error.Println("raw error sql of ", rawsql)
		return 0, fmt.Errorf("Query error: %v", err)
	}
	return count, nil
}
