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
	UserKeyTypArrayString
)

type UserKeySQLDao struct {
	db SqlLike
}

type UserKeyDao interface {
	UserKeys(delegatee SqlLike, query UserKeyQuery) ([]UserKey, error)
}

type UserKeyQuery struct {
	ID               []int64
	InputNames       []string
	Enterprise       string
	IgnoreSoftDelete bool
	paging           *Pagination
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

func NewUserKeyDao(db SqlLike) UserKeyDao {
	return &UserKeySQLDao{
		db: db,
	}
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
	if query.paging != nil {
		offsetPart = query.paging.offsetSQL()
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
