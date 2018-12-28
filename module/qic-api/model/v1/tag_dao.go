package model

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
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
	ID         []uint64
	UUID       []string
	Enterprise *string
	Name       *string
	IsDeleted  *bool
	Limit      int
	Page       int
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
	var rawsql string
	if len(conditions) > 0 {
		rawsql = "WHERE " + strings.Join(conditions, " AND ")
	}
	return rawsql, bindedData
}

func (t *TagQuery) offsetSQL() string {
	offset := t.Limit * t.Page
	return fmt.Sprintf(" LIMIT %d, %d", offset, t.Limit)
}

func (t *TagSQLDao) Begin() (*sql.Tx, error) {
	return t.db.Begin()
}

type queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

var tagSelectColumns = []string{fldTagID, fldTagUUID, fldTagIsDeleted, fldTagName,
	fldTagType, fldTagPosSen, fldTagNegSen, fldTagCreateTime, fldTagUpdateTime, fldTagEnterprise}

func (t *TagSQLDao) Tags(tx *sql.Tx, query TagQuery) ([]Tag, error) {
	var q queryer
	if tx != nil {
		q = tx
	} else if t.db != nil {
		q = t.db
	} else {
		return nil, fmt.Errorf("dao")
	}
	wheresql, data := query.whereSQL()
	rawsql := "SELECT `" + strings.Join(tagSelectColumns, "`, `") + "` FROM `" + tblTags + "` " +
		wheresql + " " + query.offsetSQL()
	logger.Error.Println("raw error sql of ", rawsql, "; data: ", data)

	rows, err := q.Query(rawsql, data...)
	if err != nil {
		logger.Error.Println("raw error sql of ", rawsql, "; data: ", data)
		return nil, fmt.Errorf("Query error: %v", err)
	}
	defer rows.Close()
	tags := make([]Tag, 0)
	for rows.Next() {
		var (
			tag Tag
			// isDeleted int8
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
		// if isDeleted == 0 {
		// 	tag.IsDeleted = false
		// } else {
		// 	tag.IsDeleted = true
		// }
		tag.PositiveSentence = posSen.String
		tag.NegativeSentence = negSen.String
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}
	return tags, nil
}

func (t *TagSQLDao) NewTags(tx *sql.Tx, tags []Tag) ([]Tag, error) {

	return nil, nil
}

func (t *TagSQLDao) DeleteTags(tx *sql.Tx, query TagQuery, isSoftDelete bool) error {
	return nil
}

func (t *TagSQLDao) CountTags(tx *sql.Tx, query TagQuery) (uint, error) {
	return 0, nil
}
