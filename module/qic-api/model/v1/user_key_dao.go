package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

const (
	tblUserKey           = "UserKey"
	fldUserKeyID         = "id"
	fldUserKeyName       = "name"
	fldUserKeyEnterprise = "enterprise"
	fldUserKeyInputName  = "inputname"
	fldUserKeyType       = "type"
	fldUserKeyIsDelete   = "is_delete"
	fldUserKeyCreateTime = "create_time"
	fldUserKeyUpdateTime = "update_time"
)

//UserKeyTyp is the constant value of the User Key type
const (
	UserKeyTypDefault int8 = iota
	UserKeyTypString
	UserKeyTypTime
	UserKeyTypNumber
	UserKeyTypArray
)

var _ UserKeyDao = &UserKeySQLDao{}

// UserKeySQLDao is the implementation of SQL operation to User Key Table.
type UserKeySQLDao struct {
	db SqlLike
}

// UserKeyDao is the abstract of the operations to User Key Table.
type UserKeyDao interface {
	NewUserKey(delegatee SqlLike, key UserKey) (UserKey, error)
	UserKeys(delegatee SqlLike, query UserKeyQuery) ([]UserKey, error)
	// Counts(delegatee SqlLike, query UserKeyQuery) (int64, error)
}

type UserKeyQuery struct {
	ID               []int64
	InputNames       []string
	Enterprise       string
	IgnoreSoftDelete bool
	Paging           *Pagination
}

func (u *UserKeyQuery) whereSQL(alias string) (string, []interface{}) {
	builder := NewWhereBuilder(andLogic, alias)
	builder.In(fldUserKeyID, int64ToWildCard(u.ID...))
	builder.In(fldUserKeyInputName, stringToWildCard(u.InputNames...))
	if u.Enterprise != "" {
		builder.Eq(fldUserKeyEnterprise, u.Enterprise)
	}
	if !u.IgnoreSoftDelete {
		builder.Eq(fldUserKeyIsDelete, u.IgnoreSoftDelete)
	}
	return builder.ParseWithWhere()
}

type UserKey struct {
	ID         int64
	Name       string
	Enterprise string
	InputName  string
	Type       int8
	IsDeleted  bool
	CreateTime int64
	UpdateTime int64
}

func NewUserKeyDao(db SqlLike) *UserKeySQLDao {
	return &UserKeySQLDao{
		db: db,
	}
}

func (u *UserKeySQLDao) NewUserKey(delegatee SqlLike, key UserKey) (UserKey, error) {
	if delegatee == nil {
		delegatee = u.db
	}
	insertCols := []string{
		fldUserKeyName, fldUserKeyEnterprise, fldUserKeyInputName,
		fldUserKeyType, fldUserKeyIsDelete, fldUserKeyCreateTime,
		fldUserKeyUpdateTime,
	}
	rawsql := fmt.Sprintf("INSERT INTO `%s`(`%s`) VALUES(? %s)",
		tblUserKey,
		strings.Join(insertCols, "`, `"),
		strings.Repeat(",? ", len(insertCols)-1),
	)
	var isDeleted int8
	if key.IsDeleted {
		isDeleted = 1
	}
	stmt, err := delegatee.Prepare(rawsql)
	if err != nil {
		return UserKey{}, fmt.Errorf("sql prepare failed, %v", err)
	}
	defer stmt.Close()
	result, err := stmt.Exec(
		key.Name, key.Enterprise, key.InputName,
		key.Type, isDeleted, key.CreateTime,
		key.UpdateTime,
	)
	if err != nil {
		return UserKey{}, fmt.Errorf("insert stmt execute failed, %v", err)
	}
	key.ID, _ = result.LastInsertId()
	return key, nil
}

func (u *UserKeySQLDao) UserKeys(delegatee SqlLike, query UserKeyQuery) ([]UserKey, error) {
	if delegatee == nil {
		delegatee = u.db
	}
	selectCols := []string{
		fldUserKeyID, fldUserKeyName, fldUserKeyEnterprise,
		fldUserKeyInputName, fldUserKeyType, fldUserKeyIsDelete,
		fldUserKeyCreateTime, fldUserKeyUpdateTime,
	}
	offsetPart := ""
	wherePart, data := query.whereSQL("")
	if query.Paging != nil {
		offsetPart = query.Paging.offsetSQL()
	}
	rawsql := fmt.Sprintf("SELECT `%s` FROM `%s` %s ORDER BY `%s` %s",
		strings.Join(selectCols, "`,`"),
		tblUserKey,
		wherePart,
		fldUserKeyCreateTime,
		offsetPart,
	)
	rows, err := delegatee.Query(rawsql, data...)
	if err != nil {
		logger.Error.Println("error rawsql, ", rawsql)
		return nil, fmt.Errorf("sql query failed, %v", err)
	}
	defer rows.Close()
	var userKeys []UserKey
	for rows.Next() {
		var key UserKey
		var isDelete int8
		rows.Scan(
			&key.ID, &key.Name, &key.Enterprise,
			&key.InputName, &key.Type, &isDelete,
			&key.CreateTime, &key.UpdateTime,
		)
		if isDelete != 0 {
			key.IsDeleted = true
		}
		userKeys = append(userKeys, key)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan error, %v", err)
	}
	return userKeys, nil
}

// CountUserKeys return the total number of the UserKeys.
func (u *UserKeySQLDao) CountUserKeys(delegatee SqlLike, query UserKeyQuery) (int64, error) {
	if delegatee == nil {
		delegatee = u.db
	}
	wherepart, data := query.whereSQL("")
	rawsql := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s",
		tblUserKey,
		wherepart,
	)
	var total int64
	err := delegatee.QueryRow(rawsql, data...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query row failed, %v", err)
	}
	return total, nil
}
