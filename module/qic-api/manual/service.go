package manual

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
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
