package model

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

type TaskDao interface {
	CallTask(delegatee SqlLike, call Call) (Task, error)
	NewTask(delegatee SqlLike, task Task) (*Task, error)
}

type TaskSQLDao struct {
	db *sql.DB
}

func NewTaskDao(db *sql.DB) *TaskSQLDao {
	return &TaskSQLDao{
		db: db,
	}
}

// Task represent the task table of the QISYS database.
//	CallsOfStaffs is the virutal field for relationship with calls.
//	which its key is StaffID.
type Task struct {
	ID            int64
	Status        int8
	Description   string
	IsDeal        bool
	Series        string //序號
	CreatedTime   int64
	UpdatedTime   int64
	CallsOfStaffs map[string][]Call
	// Creator     string
	// Updator     string
}

// NewTask insert task into db with the its fields.
func (t *TaskSQLDao) NewTask(delegatee SqlLike, task Task) (*Task, error) {
	if delegatee == nil {
		delegatee = t.db
	}

	insertCols := []string{
		fldTaskID, fldTaskStatus, fldTaskDescription,
		fldTaskDeal, fldTaskSeries, fldTaskCreateTime,
		fldTaskUpdateTime, "creator", "updator", //SINCE WE DONT KNOW WHAT TO FILL WITH THE CREATOR OR UPDATOR, WE JUST FILL AN EMPTY STRING.
	}
	rawquery := fmt.Sprintf("INSERT INTO `%s`(`%s`) VALUE (? %s)",
		tblTask, strings.Join(insertCols, "`, `"), strings.Repeat(",?", len(insertCols)-1),
	)
	result, err := delegatee.Exec(rawquery,
		task.ID, task.Status, task.Description,
		task.IsDeal, task.Series, task.CreatedTime,
		task.UpdatedTime, "", "")
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

//TaskQuery is the query conditions of task table.
type TaskQuery struct {
	ID []int64
	SN []string // Serial Number
}

func (q TaskQuery) whereSQL() (string, []interface{}) {
	var (
		whereBuilder = NewWhereBuilder(andLogic, "")
	)
	whereBuilder.In(fldTaskID, int64ToWildCard(q.ID...))
	whereBuilder.In(fldTaskSeries, stringToWildCard(q.SN...))
	return whereBuilder.ParseWithWhere()
}

// CallTask query task by the given call.
func (t *TaskSQLDao) CallTask(delegatee SqlLike, call Call) (Task, error) {
	tasks, err := t.Task(delegatee, TaskQuery{ID: []int64{call.TaskID}})
	if err != nil {
		return Task{}, err
	} else if len(tasks) == 0 {
		return Task{}, fmt.Errorf("no such task")
	}
	return tasks[0], err
}

// Task query the task by the given query.
// returned []Task are ordered by id(create time) ascending.
func (t *TaskSQLDao) Task(delegatee SqlLike, query TaskQuery) ([]Task, error) {
	if delegatee == nil {
		delegatee = t.db
	}
	selectCols := []string{
		fldTaskID, fldTaskStatus, fldTaskDescription,
		fldTaskDeal, fldTaskSeries, fldTaskCreateTime,
		fldTaskUpdateTime,
	}
	wherePart, data := query.whereSQL()

	rawquery := fmt.Sprintf("SELECT `%s` FROM `%s` %s ORDER BY `%s` ASC",
		strings.Join(selectCols, "`, `"), tblTask, wherePart, fldTaskID,
	)
	rows, err := delegatee.Query(rawquery, data...)
	if err != nil {
		logger.Error.Println("raw error sql: ", rawquery)
		return nil, fmt.Errorf("sql query failed, %v", err)
	}
	defer rows.Close()
	var tasks = make([]Task, 0)
	for rows.Next() {
		var (
			task        Task
			isDeleted   int8
			description sql.NullString
		)
		rows.Scan(
			&task.ID, &task.Status, &description,
			&isDeleted, &task.Series, &task.CreatedTime,
			&task.UpdatedTime,
		)
		task.IsDeal = (isDeleted != 0)
		task.Description = description.String
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}
	return tasks, nil
}
