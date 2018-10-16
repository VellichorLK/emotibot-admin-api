package Gray

import (
	"strings"
	"fmt"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func QueryTotalWhite(condition QueryCondition, appid string)(int, error) {
	query := "select ifnull(count(1), 0) as totalCount from white_list where is_deleted=0"
	var total int
	db := util.GetAuditDB()
	if db == nil {
		return 0, fmt.Errorf("DB not init")
	}

	// fetch
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return total, err;
}

func FetchWhites(condition QueryCondition, appid string) ([]White, error) {
	var whites []White
	db := util.GetAuditDB()
	if db == nil {
		return nil, fmt.Errorf("DB not init")
	}

	query := "select user_id as UserId from white_list where is_deleted = 0 order by create_time desc "

	// fetch
	rows, err := db.Query(query)
	if err != nil {
		return whites, err
	}
	defer rows.Close()

	for rows.Next() {
		var white White
		rows.Scan(&white.UserId)
		whites = append(whites, white)
	}

	return whites, nil
}

func BatchInsertWhite(userId string, appid string) (int, error) {
	db := util.GetAuditDB()
	if db == nil {
		return 0, fmt.Errorf("DB not init")
	}

	insertSql := "insert into white_list(user_id,is_deleted) values (?, 0) "
	insertSmtm, err := db.Prepare(insertSql)
	if err != nil {
		return 0, err
	}
	begin , err := db.Begin()
	if err != nil {
		return 0, err
	}
	userIds := strings.Split(userId, ",")
	for i:=0; i<len(userIds); i++ {
		_, err := insertSmtm.Exec(userIds[i])
		if err != nil {
			defer insertSmtm.Close()
			defer begin.Rollback()
            return 0, err
        }
	}
	defer insertSmtm.Close()
	err = begin.Commit()
	if err != nil {
		defer begin.Rollback()
		return 0, err
	}

	return len(userIds), nil
}

func BatchDeleteWhite(userId string, appid string) (int, error) {
	db := util.GetAuditDB()
	if db == nil {
		return 0, fmt.Errorf("DB not init")
	}

	delSql := "update white_list set is_deleted=1 where user_id=? "
	delSmtm, err := db.Prepare(delSql)
	if err != nil {
		return 0, err
	}
	begin , err := db.Begin()
	if err != nil {
		return 0, err
	}
	userIds := strings.Split(userId, ",")
	for i:=0; i<len(userIds); i++ {
		_, err := delSmtm.Exec(userIds[i])
		if err != nil {
			defer delSmtm.Close()
			defer begin.Rollback()
            return 0, err
        }
	}
	defer delSmtm.Close()
	err = begin.Commit()
	if err != nil {
		defer begin.Rollback()
		return 0, err
	}

	return len(userIds), nil
}