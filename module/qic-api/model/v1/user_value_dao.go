package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

// UserValueDao is the implementation of SQL operation to User Value Table.
type UserValueDao struct {
	conn SqlLike
}

// UserValue is the value of custom column in database.
type UserValue struct {
	ID         int64
	UserKeyID  int64
	LinkID     int64 // LinkID is parent table id
	Type       int8  // Type is different type point to different table
	Value      string
	IsDeleted  bool
	CreateTime int64
	UpdateTime int64
	UserKey    *UserKey
}

// the enum of the UserValue Type
const (
	UserValueTypDefault int8 = iota
	UserValueTypGroup
	UserValueTypCall
	UserValueTypSensitiveWord
)

const (
	tblUserValue = "UserValue"
)

const (
	fldUserValueID         = "id"
	fldUserValueUserKey    = "userkey_id"
	fldUserValueLinkID     = "link_id"
	fldUserValueType       = "type"
	fldUserValueVal        = "value"
	fldUserValueIsDelete   = "is_delete"
	fldUserValueCreateTime = "create_time"
	fldUserValueUpdateTime = "update_time"
)

// UserValueQuery is the query for UserValue table
type UserValueQuery struct {
	ID               []int64
	UserKeyID        []int64
	Type             []int8
	IgnoreSoftDelete bool
	Paging           *Pagination
}

func (u *UserValueQuery) whereSQL(alias string) (string, []interface{}) {
	return u.whereBuilder(alias).ParseWithWhere()
}

func (u *UserValueQuery) whereBuilder(alias string) *whereBuilder {
	builder := NewWhereBuilder(andLogic, alias)
	builder.In(fldUserValueID, int64ToWildCard(u.ID...))
	builder.In(fldUserValueUserKey, int64ToWildCard(u.UserKeyID...))
	builder.In(fldUserValueType, int8ToWildCard(u.Type...))
	if !u.IgnoreSoftDelete {
		builder.Eq(fldUserKeyIsDelete, false)
	}
	return builder
}

// UserValues search the user tables and find by it given query.
// Since values almost don't used without it key.
// Use ValuesKey or Key's KeyValues function will be recommend.
func (u *UserValueDao) UserValues(delegatee SqlLike, query UserValueQuery) ([]UserValue, error) {
	if delegatee == nil {
		delegatee = u.conn
	}
	selectCols := []string{
		fldUserValueID, fldUserValueUserKey, fldUserValueLinkID,
		fldUserValueType, fldUserValueVal, fldUserValueIsDelete,
		fldUserValueCreateTime, fldUserValueUpdateTime,
	}
	wherePart, data := query.whereSQL("")
	var offsetPart string
	if query.Paging != nil {
		offsetPart = query.Paging.offsetSQL()
	}
	rawsql := fmt.Sprintf("SELECT `%s` FROM `%s` %s ORDER BY `%s` DESC %s",
		strings.Join(selectCols, "`, `"),
		tblUserValue,
		wherePart,
		fldUserValueCreateTime,
		offsetPart,
	)
	rows, err := delegatee.Query(rawsql, data...)
	if err != nil {
		logger.Error.Println("raw error sql: ", rawsql)
		return nil, fmt.Errorf("sql query failed, %v", err)
	}
	defer rows.Close()
	userValues := []UserValue{}
	for rows.Next() {
		var (
			uv        UserValue
			isDeleted int8
		)
		rows.Scan(
			&uv.ID, &uv.UserKeyID, &uv.LinkID,
			&uv.Type, &uv.Value, &isDeleted,
			&uv.CreateTime, &uv.UpdateTime,
		)
		if isDeleted != 0 {
			uv.IsDeleted = true
		}
		userValues = append(userValues, uv)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return userValues, nil
}

// NewUserValue insert a new user value.
func (u *UserValueDao) NewUserValue(delegatee SqlLike, user UserValue) (UserValue, error) {
	if delegatee == nil {
		delegatee = u.conn
	}

	insertCols := []string{
		fldUserValueUserKey, fldUserValueLinkID, fldUserValueType,
		fldUserValueVal, fldUserValueIsDelete, fldUserValueCreateTime,
		fldUserValueUpdateTime,
	}

	rawsql := fmt.Sprintf("INSERT INTO `%s`(`%s`) VALUES(?%s)",
		tblUserValue,
		strings.Join(insertCols, "`, `"),
		strings.Repeat(",?", len(insertCols)-1),
	)

	result, err := delegatee.Exec(rawsql,
		user.UserKeyID, user.LinkID, user.Type,
		user.Value, user.IsDeleted, user.CreateTime,
		user.UpdateTime,
	)
	if err != nil {
		logger.Error.Println("raw error sql: ", rawsql)
		return UserValue{}, fmt.Errorf("execute insert failed, %v", err)
	}
	user.ID, err = result.LastInsertId()
	if err != nil {
		return UserValue{}, ErrAutoIDDisabled
	}
	return user, nil
}

// DeleteUserValues is a soft delete operation, which mark the query values as deleted.
func (u *UserValueDao) DeleteUserValues(delegatee SqlLike, query UserValueQuery) (int64, error) {
	if delegatee == nil {
		delegatee = u.conn
	}
	wherePart, data := query.whereSQL("")
	rawsql := fmt.Sprintf("UPDATE `%s` SET `%s` = 1 %s",
		tblUserValue, fldUserValueIsDelete, wherePart,
	)
	result, err := delegatee.Exec(rawsql, data...)
	if err != nil {
		return 0, fmt.Errorf("execute delete failed, %v", err)
	}
	total, err := result.RowsAffected()
	if err != nil {
		return 0, ErrAutoIDDisabled
	}
	return total, nil
}

// ValuesKey get value and its key.
// all UserValue's key returned should be populated
func (u *UserValueDao) ValuesKey(delegatee SqlLike, query UserValueQuery) ([]UserValue, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED YET")
}