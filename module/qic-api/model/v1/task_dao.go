package model

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

type TaskDao struct {
	db *sql.DB
}

type Task struct {
	ID     int64
	Status int8
	// Description string
	IsDeal bool
	//序號
	Series        string
	CreatedTime   int64
	UpdatedTime   int64
	CallsOfStaffs map[string][]Call
	// Creator     string
	// Updator     string
}

func (t *TaskDao) NewTask(delegatee SqlLike, task Task) (*Task, error) {
	if delegatee == nil {
		delegatee = t.db
	}
	//SINCE WE DONT KNOW WHAT TO FILL THE CREATOR OR UPDATOR, WE JUST FILL AN EMPTY STRING.
	insertCols := []string{fldTaskID, fldTaskStatus, fldTaskDeal,
		fldTaskSeries, fldTaskCreateTime, fldTaskUpdateTime, "creator", "updator"}
	rawquery := "INSERT INTO `" + tblTask + "`(`" + strings.Join(insertCols, "`, `") + "`) VALUE (?" + strings.Repeat(",?", len(insertCols)-1) + ")"
	result, err := delegatee.Exec(rawquery, task.ID, task.Status,
		task.IsDeal, task.Series, task.CreatedTime, task.UpdatedTime, "", "")
	if err != nil {
		logger.Error.Println("raw error sql: ", rawquery)
		return nil, fmt.Errorf("sql execute failed, %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, ErrAutoIDDisabled
	}
	task.ID = id
	return &task, nil
}

type TaskQuery struct {
	ID []int64
}

func (q TaskQuery) whereSQL() (string, []interface{}) {
	var (
		conditions []string
		bindData   = make([]interface{}, 0)
		rawCond    string
	)
	if len(q.ID) > 0 {
		cond := fldTaskID + " IN (?" + strings.Repeat(",? ", len(q.ID)-1) + ")"
		conditions = append(conditions, cond)
		for _, id := range q.ID {
			bindData = append(bindData, id)
		}
	}
	if len(conditions) == 0 {
		return "", bindData
	}
	rawCond = " WHERE " + strings.Join(conditions, " AND ")
	return rawCond, bindData
}

func (t *TaskDao) CallTask(delegatee SqlLike, call Call) (Task, error) {
	tasks, err := t.Task(delegatee, TaskQuery{ID: []int64{call.TaskID}})
	if err != nil {
		return Task{}, err
	} else if len(tasks) == 0 {
		return Task{}, fmt.Errorf("no such task")
	}
	return tasks[0], err
}

func (t *TaskDao) Task(delegatee SqlLike, query TaskQuery) ([]Task, error) {
	if delegatee == nil {
		delegatee = t.db
	}
	selectCols := []string{fldTaskID, fldTaskStatus, fldTaskDeal,
		fldTaskSeries, fldTaskCreateTime, fldTaskUpdateTime}
	wherePart, data := query.whereSQL()
	rawquery := "SELECT `" + strings.Join(selectCols, "`, `") + "` FROM `" + tblTask + "` " + wherePart
	rows, err := delegatee.Query(rawquery, data...)
	if err != nil {
		logger.Error.Println("raw error sql: ", rawquery)
		return nil, fmt.Errorf("sql query failed, %v", err)
	}
	defer rows.Close()
	var tasks = make([]Task, 0)
	for rows.Next() {
		var task Task
		rows.Scan(&task.ID, &task.Status, &task.IsDeal, &task.Series, &task.CreatedTime, &task.UpdatedTime)
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}
	return tasks, nil
}
