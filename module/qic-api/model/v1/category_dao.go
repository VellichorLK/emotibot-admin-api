package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

//CategoryDao defines the db access function
type CategoryDao interface {
	GetCategories(conn SqlLike, q *CategoryQuery) ([]*CategortInfo, error)
	InsertCategory(conn SqlLike, s *CategoryRequest) (int64, error)
	SoftDeleteCategory(conn SqlLike, q *CategoryQuery) (int64, error)
	CountCategory(conn SqlLike, q *CategoryQuery) (uint64, error)
	UpdateCategory(conn SqlLike, id uint64, s *CategoryRequest) error
}

//CategorySQLDao is the structure to implement the db access
type CategorySQLDao struct {
}

//CategoryRequest is the request of new category
type CategoryRequest struct {
	Name       string `json:"name"`
	Type   int8
	Enterprise string
}

//CategortInfo stores the category information used to return
type CategortInfo struct {
	ID         uint64 `json:"category_id,string"`
	Name       string `json:"name"`
	Enterprise string `json:"-"`
}

//CategoryQuery uses as query structure of category
type CategoryQuery struct {
	ID         []uint64
	Enterprise *string
	Page       int
	Limit      int
	IsDelete   *int8
	Type *int8
}

func (q *CategoryQuery) whereSQL() (string, []interface{}) {
	numOfID := len(q.ID)
	params := make([]interface{}, 0, numOfID+1)
	conditions := []string{}

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

	if q.Type != nil {
		condition := fldType + " = ?"
		conditions = append(conditions, condition)
		params = append(params, *q.Type)
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

//GetCategories gets the categories
func (c *CategorySQLDao) GetCategories(conn SqlLike, q *CategoryQuery) ([]*CategortInfo, error) {
	if conn == nil {
		return nil, ErrNilSqlLike
	}
	if q == nil {
		return nil, ErrNeedCondition
	}
	condition, params := q.whereSQL()

	querySQL := fmt.Sprintf("SELECT %s,%s,%s FROM %s %s",
		fldID, fldName, fldEnterprise,
		tblCategory,
		condition)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s, %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()

	infos := make([]*CategortInfo, 0)
	for rows.Next() {
		var c CategortInfo
		err = rows.Scan(&c.ID, &c.Name, &c.Enterprise)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		infos = append(infos, &c)
	}
	return infos, nil
}

//InsertCategory inserts a new category
func (c *CategorySQLDao) InsertCategory(conn SqlLike, r *CategoryRequest) (int64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	if r == nil {
		return 0, ErrNeedCondition
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s,%s, %s) VAlUES (?,?,?)", tblCategory,
		fldName, fldEnterprise, fldType)
	res, err := conn.Exec(insertSQL, r.Name, r.Enterprise, r.Type)
	if err != nil {
		logger.Error.Printf("Exec sql failed. %s [%s %s]\n", insertSQL, r.Name, r.Enterprise)
		return 0, err
	}
	return res.LastInsertId()
}

//SoftDeleteCategory delete the category by setting is_delete to 1
func (c *CategorySQLDao) SoftDeleteCategory(conn SqlLike, q *CategoryQuery) (int64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	if q == nil {
		return 0, ErrNeedCondition
	}

	condition, params := q.whereSQL()

	deleteSQL := fmt.Sprintf("Update %s SET %s=1 %s", tblCategory, fldIsDelete, condition)

	res, err := conn.Exec(deleteSQL, params...)
	if err != nil {
		logger.Error.Printf("delete failed. %s. %s [%+v]\n", err, deleteSQL, params)
		return 0, err
	}
	return res.RowsAffected()
}

//CountCategory counts number of categories that meets the condition
func (c *CategorySQLDao) CountCategory(conn SqlLike, q *CategoryQuery) (uint64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	if q == nil {
		return 0, ErrNeedCondition
	}

	condition, params := q.whereSQL()

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tblCategory, condition)

	var count uint64
	err := conn.QueryRow(countSQL, params...).Scan(&count)
	if err != nil {
		logger.Error.Printf("count failed. %s\n %s %+v\n", err, countSQL, params)
		return 0, err
	}
	return count, nil
}

// UpdateCategory updates category name in which meets the id
func (c *CategorySQLDao) UpdateCategory(conn SqlLike, id uint64, r *CategoryRequest) (err error) {
	if conn == nil {
		return ErrNilSqlLike
	}

	if r == nil {
		return ErrNeedCondition
	}

	updateStr := fmt.Sprintf(
		"UPDATE %s SET %s=? WHERE %s=?",
		tblCategory,
		fldName,
		fldID,
	)

	_, err = conn.Exec(updateStr, r.Name, id)
	if err != nil {
		logger.Error.Printf("udpate failed in UpdateCategory, %s\n %s %+v\n", err, updateStr, r)
		return
	}
	return
}
