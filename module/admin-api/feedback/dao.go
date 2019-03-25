package feedback

import (
	"database/sql"
	"errors"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

// Dao is the interface of feedback dao, it can be used for mock
type Dao interface {
	GetReasons(appid string) ([]*Reason, error)
	AddReason(appid string, content string) (int64, error)
	UpdateReason(appid string, id int64, content string) error
	DeleteReason(appid string, id int64) error
}

type feedbackDao struct {
	db *sql.DB
}

var (
	// ErrDBNotInit is used to be returned if dao is not initialized
	ErrDBNotInit = errors.New("DB is not init")
	// ErrIDNotExisted is used when id is not existed in update
	ErrIDNotExisted = errors.New("ID is not found")
	// ErrDuplicateContent is used when trying to insert duplicate reason
	ErrDuplicateContent = errors.New("Duplicate content")

	// it is used for mock
	timestampHandler = func() int64 {
		return time.Now().Unix()
	}
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
	idx := 0
	for rows.Next() {
		t := Reason{}
		err = rows.Scan(&t.ID, &t.Content)
		if err != nil {
			return nil, err
		}
		t.Index = idx
		idx++
		ret = append(ret, &t)
	}

	return ret, nil
}

func (dao feedbackDao) AddReason(appid string, content string) (int64, error) {
	if dao.db == nil {
		return 0, ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return 0, err
	}
	defer util.ClearTransition(tx)

	queryStr := "SELECT 1 FROM feedback_reason WHERE appid = ? AND content = ?"
	row := tx.QueryRow(queryStr, appid, content)
	ret := 0
	err = row.Scan(&ret)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	err = nil
	if ret == 1 {
		logger.Error.Printf("Trying to add duplicate reason with [%s] [%s]\n", appid, content)
		return 0, ErrDuplicateContent
	}

	queryStr = `
		INSERT INTO feedback_reason
			(appid, content, created_at)
			VALUES (?, ?, ?)
		`
	result, err := dao.db.Exec(queryStr, appid, content, timestampHandler())
	if err != nil {
		return 0, err
	}

	_, err = result.RowsAffected()
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

	result, err := dao.db.Exec(sql, appid, id)

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrIDNotExisted
	}

	if err != nil {
		return err
	}

	return nil
}

func (dao feedbackDao) UpdateReason(appid string, id int64, content string) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	sql := "UPDATE feedback_reason SET content = ? WHERE appid = ? AND id = ?"
	result, err := dao.db.Exec(sql, content, appid, id)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrIDNotExisted
	}

	return nil
}
