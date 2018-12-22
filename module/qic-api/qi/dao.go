package qi

import (
	"fmt"
	"database/sql"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

type DAO interface {
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	ClearTranscation(tx *sql.Tx)
	GetGroups() ([]Group, error)
	CreateGroup(group *Group, tx *sql.Tx) (*Group, error)
	GetGroupBy(id int64) (*Group, error)
	UpdateGroup(id int64, group *Group, tx *sql.Tx) (error)
	DeleteGroup(id int64) (error)
}

type sqlDAO struct {
	conn *sql.DB
}

//InitDB is used to get the db in this module
func (s *sqlDAO) initDB() error {
	if s.conn == nil {
		envs := ModuleInfo.Environments

		url := envs["MYSQL_URL"]
		user := envs["MYSQL_USER"]
		pass := envs["MYSQL_PASS"]
		db := envs["MYSQL_DB"]

		conn, err := util.InitDB(url, user, pass, db)
		if err != nil {
			logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
			return err
		}
		s.conn = conn
	}
	return nil
}

//Begin is used to start a transaction
func (s *sqlDAO) Begin() (*sql.Tx, error) {
	if s.conn == nil {
		err := s.initDB()
		if err != nil {
			return nil, err
		}
	}
	return s.conn.Begin()
}

//Commit commits the data
func (s *sqlDAO) Commit(tx *sql.Tx) error {
	if tx != nil {
		return tx.Commit()
	}
	return nil
}

func (s *sqlDAO) ClearTranscation(tx *sql.Tx) {
	if tx != nil {
		util.ClearTransition(tx)
	}
}

func (s *sqlDAO) GetGroups() (groups []Group, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			return
		}
	}

	queryStr := "SELECT id, group_name FROM rule_group where `is_enable`=1"

	rows, err := s.conn.Query(queryStr)
	if err != nil {
		err = fmt.Errorf("error while query groups in dao.GetGroups, err: %s", err.Error())
		return
	}
	defer rows.Close()

	groups = make([]Group, 0)
	for rows.Next() {
		group := Group{}
		rows.Scan(&group.ID, &group.Name)

		groups = append(groups, group)
	}
	return
}

func genInsertRelationSQL(id int64, rules []int64) (str string, values []interface{}) {
	str = "INSERT INTO Relation_Group_Rule (group_id, rule_id) VALUES "
	values = []interface{}{}
	for _, ruleID := range rules {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " (?, ?)"
		values = append(values, id, ruleID)
	}
	return
}

func (s *sqlDAO) CreateGroup(group *Group, tx *sql.Tx) (createdGroup *Group, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	now := time.Now().Unix()

	// insert group
	insertStr := "INSERT INTO `rule_group` (group_name, enterprise, create_time, update_time, is_enable, limit_speed, limit_silence) VALUES (?, ?, ?, ?, ?, ?, ?)"
	values := []interface{}{
		group.Name,
		group.Enterprise,
		now,
		now,
		group.Enabled,
		group.Speed,
		group.SlienceDuration,
	}
	result, err := tx.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert group in dao.CreateGroup, err: %s", err.Error())
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get group id in dao.CreateGroup, err: %s", err.Error())
		return
	}

	// insert condition
	insertStr = "INSERT INTO group_condition (group_id, file_name, deal, series, staff_id, staff_name, extension, department, client_id, client_name, client_phone, call_start, call_end, left_channel, right_channel) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	values = []interface{}{
		groupID,
		group.Condition.FileName,
		group.Condition.Deal,
		group.Condition.Series,
		group.Condition.StaffID,
		group.Condition.StaffName,
		group.Condition.Extension,
		group.Condition.Department,
		group.Condition.ClientID,
		group.Condition.ClientName,
		group.Condition.ClientPhone,
		group.Condition.CallStart,
		group.Condition.CallEnd,
		group.Condition.LeftChannelCode,
		group.Condition.RightChannelCode,
	}

	_, err = tx.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert condition in dao.CreateGroup, err: %s", err.Error())
		return
	}

	// insert into group_rule_map
	if len(group.Rules) != 0 {
		insertStr, values = genInsertRelationSQL(groupID, group.Rules)

		_, err = tx.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert relation_group_rule in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	group.ID = groupID
	createdGroup = group
	return
}

func (s *sqlDAO) GetGroupBy(id int64) (group *Group, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	queryStr := `SELECT g.id, g.group_name, g.limit_speed, g.limit_silence, 
	gc.file_name, gc.deal, gc.series, gc.staff_id, gc.staff_name, gc.extension, gc.department, 
	gc.client_id, gc.client_name, gc.client_phone, gc.call_start, gc.call_end, gc.left_channel, gc.right_channel
	FROM (SELECT * FROM rule_group WHERE id=?) as g 
	LEFT JOIN group_condition as gc ON g.id = gc.group_id`

	rows, err := s.conn.Query(queryStr, id)
	if err != nil {
		err = fmt.Errorf("error while query group in dao.GetGroupBy, err: %s", err.Error())
		return
	}

	group = &Group{}

	for rows.Next() {
		condition := GroupCondition{}

		rows.Scan(
			&group.ID,
			&group.Name,
			&group.Speed,
			&group.SlienceDuration,
			&condition.FileName,
			&condition.Deal,
			&condition.Series,
			&condition.StaffID,
			&condition.StaffName,
			&condition.Extension,
			&condition.Department,
			&condition.ClientID,
			&condition.ClientName,
			&condition.ClientPhone,
			&condition.CallStart,
			&condition.CallEnd,
			&condition.LeftChannelCode,
			&condition.RightChannelCode,
		)
		group.Condition = &condition
	}

	// get rules under this group
	queryStr = "SELECT rule_id FROM Relation_Group_Rule WHERE group_id = ?"
	rows, err = s.conn.Query(queryStr, id)
	if err != nil {
		err = fmt.Errorf("error while get rules of group in dao.GetGroupBy, err: %s", err.Error())
		return
	}

	group.Rules = make([]int64, 0)
	for rows.Next() {
		var ruleID int64
		rows.Scan(&ruleID)
		group.Rules = append(group.Rules, ruleID)
	}

	return
}

func (s *sqlDAO) UpdateGroup(id int64, group *Group, tx *sql.Tx) (err error) {
	if group == nil {
		return
	}

	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	// update group
	updateStr, values := genUpdateGroupSQL(id, group)
	_, err = tx.Exec(updateStr, values...)
	if err != nil {
		err = fmt.Errorf("error while update group in dao.UpdateGroup, err: %s", err.Error())
		return
	}

	// update condition
	if group.Condition != nil {
		updateStr, values = genUpdateConditionSQL(id, group.Condition)
		_, err = tx.Exec(updateStr, values...)
		if err != nil {
			err = fmt.Errorf("error while update condition in dao.UpdateGroup, err: %s", err.Error())
		}
	}

	// update relation
	// delete old relation 
	// add new relation
	updateStr = "DELETE FROM Relation_Group_Rule WHERE group_id=?"
	_, err = tx.Exec(updateStr, id)
	if err != nil {
		err = fmt.Errorf("error while delete old relation in dao.UpdateGroup, err: %s", err.Error())
		return
	}

	if len(group.Rules) != 0 {
		updateStr, values = genInsertRelationSQL(id, group.Rules)
		_, err = tx.Exec(updateStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert new relation in dao.UpdateGroup, err: %s", err.Error())
			return
		}
	}
	return
}

func genUpdateGroupSQL(id int64, group *Group) (str string, values []interface{}) {
	str = "UPDATE rule_group SET "

	values = make([]interface{}, 0)
	if group.Name != "" {
		str += "group_name = ?"
		values = append(values, group.Name)
	}

	if group.Speed != 0 {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " limit_speed = ?"
		values = append(values, group.Speed)
	}

	if group.SlienceDuration != 0 {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " limit_silence = ?"
		values = append(values, group.SlienceDuration)
	}

	if group.Enterprise != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " enterprise = ?"
		values = append(values, group.Enterprise)
	}

	str = addCommaIfNotFirst(str, len(values) == 0)
	str += " is_enable = ?"
	values = append(values, group.Enabled)

	str = fmt.Sprintf("%s where id = ?", str)
	values = append(values, id)
	return
}

func genUpdateConditionSQL(id int64, condition *GroupCondition) (str string , values []interface{}) {
	str = "UPDATE group_condition SET "
	values = make([]interface{}, 0)

	if condition.CallEnd != 0 {
		str += "call_end = ?"
		values = append(values, condition.CallEnd)
	}

	if condition.CallStart != 0 {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " call_start = ?"
		values = append(values, condition.CallStart)
	}

	if condition.ClientID != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " client_id = ?"
		values = append(values, condition.ClientID)
	}

	if condition.ClientName != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " client_name = ?"
		values = append(values, condition.ClientName)
	}

	if condition.ClientPhone != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " client_phone = ?"
		values = append(values, condition.ClientPhone)
	}

	if condition.Department != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " department = ?"
		values = append(values, condition.Department)
	}

	if condition.Extension != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " extension = ?"
		values = append(values, condition.Extension)
	}

	if condition.FileName != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " file_name = ?"
		values = append(values, condition.FileName)
	}

	if condition.Series != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " series = ?"
		values = append(values, condition.Series)
	}

	if condition.StaffID != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " staff_id = ?"
		values = append(values, condition.StaffID)
	}

	if condition.StaffName != "" {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " staff_name = ?"
		values = append(values, condition.StaffName)
	}

	str = addCommaIfNotFirst(str, len(values) == 0)
	str += " deal = ?, left_channel = ?, right_channel = ? where group_id = ?"
	values = append(values, condition.Deal, condition.LeftChannelCode, condition.RightChannelCode, id)
	return
}

func addCommaIfNotFirst(sqlStr string, first bool) (string) {
	if !first {
		sqlStr += ","
		return sqlStr
	}
	return sqlStr
}

func (s *sqlDAO) DeleteGroup(id int64) (err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	deleteStr := "UPDATE rule_group SET is_delete = 1 WHERE id = ?"
	_, err = s.conn.Exec(deleteStr, id)
	if err != nil {
		err = fmt.Errorf("error while delete group in dao.DeleteGroup, err: %s", err.Error())
	}
	return
}