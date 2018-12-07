package feedback

import (
	"database/sql"
	"errors"
	"time"
)

// Dao is the interface of feedback dao, it can be used for mock
type Dao interface {
	GetReasons(appid string) ([]*Reason, error)
	AddReason(appid string, content string) (int64, error)
	DeleteReason(appid string, id int64) error
}

type feedbackDao struct {
	db *sql.DB
}

var (
	// ErrDBNotInit is used to be returned if dao is not initialized
	ErrDBNotInit = errors.New("DB is not init")
	// it is used for mock
	timestampHandler = getTimestamp
)

func (dao feedbackDao) GetReasons(appid string) ([]*Reason, error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	sql := "SELECT id, content FROM feedback_reason WHERE appid = ? ORDER BY id"
	rows, err := dao.db.Query(sql, appid)
	if err != nil {
		return nil, err
	}

	ret := []*Reason{}
	for rows.Next() {
		t := Reason{}
		err = rows.Scan(&t.ID, &t.Content)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &t)
	}

	return ret, nil
}

func (dao feedbackDao) AddReason(appid string, content string) (int64, error) {
	if dao.db == nil {
		return 0, ErrDBNotInit
	}

	sql := "INSERT INTO feedback_reason (appid, content, created_time) VALUES (?, ?, ?)"

	result, err := dao.db.Exec(sql, appid, content, timestampHandler())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (dao feedbackDao) DeleteReason(appid string, id int64) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	sql := "DELETE FROM feedback_reason WHERE appid = ? AND id = ?"

	_, err := dao.db.Exec(sql, appid, id)
	if err != nil {
		return err
	}

	return nil
}

func getTimestamp() int64 {
	return time.Now().Unix()
}
