package manual

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	_ "emotibot.com/emotigo/pkg/logger"
	"fmt"
)

var (
	ErrNilTask = fmt.Errorf("Nil InspectorTask Error")
)

var taskDao model.InspectTaskDao = &model.InspectTaskSqlDao{}

func CreateTask(task *model.InspectTask) (uuid string, err error) {
	if task == nil {
		err = ErrNilTask
		return
	}

	// create uuid for the new flow
	uuid, err = general.UUID()
	if err != nil {
		err = fmt.Errorf("error while create uuid in CreateTask, err: %s", err.Error())
		return
	}

	task.UUID = uuid

	tx, err := manualDB.Begin()
	if err != nil {
		return
	}
	defer manualDB.ClearTransition(tx)

	_, err = taskDao.Create(task, tx)
	if err != nil {
		return
	}
	err = manualDB.Commit(tx)
	return
}

func GetTasks(filter *model.InspectTaskFilter) (total int64, tasks []model.InspectTask, err error) {
	manualConn := manualDB.Conn()

	total, err = taskDao.CountBy(filter, manualConn)
	if err != nil {
		return
	}

	tasks, err = taskDao.GetBy(filter, manualConn)
	if err != nil {
		return
	}

	userIDs := make([]string, len(tasks))
	for idx, task := range tasks {
		userIDs[idx] = task.Creator
	}

	authConn := authDB.Conn()
	usersMap, err := taskDao.Users(userIDs, authConn)
	if err != nil {
		return
	}

	for idx := range tasks {
		task := &tasks[idx]
		task.Creator = usersMap[task.Creator]
	}
	return
}
