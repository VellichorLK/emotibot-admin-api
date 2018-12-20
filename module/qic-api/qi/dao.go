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

	queryStr := "SELECT id, group_name FROM `Group` where `is_enable`=1"

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
	insertStr := "INSERT INTO `Group` (group_name, enterprise, create_time, update_time, is_enable, limit_speed, limit_silence) VALUES (?, ?, ?, ?, ?, ?, ?)"
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
	insertStr = "INSERT INTO `Condition` (group_id, file_name, deal, series, staff_id, staff_name, extension, department, client_id, client_name, client_phone, call_start, call_end, left_channel, right_channel) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

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

	group.ID = groupID
	createdGroup = group
	return
}