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

func CreateTask(task *model.InspectTask) (id int64, err error) {
	if task == nil {
		err = ErrNilTask
		return
	}

	// create uuid for the new flow
	uuid, err := general.UUID()
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

	id, err = taskDao.Create(task, tx)
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

	taskIDs := make([]int64, len(tasks))
	for idx, task := range tasks {
		taskIDs[idx] = task.ID
	}

	taskFilter := &model.StaffTaskFilter{
		TaskIDs: taskIDs,
	}

	tasksInfo, err := taskDao.TasksInfoBy(taskFilter, manualConn)
	if err != nil {
		return
	}

	userIDExists := map[string]bool{}
	userIDs := []string{}
	for _, task := range tasks {
		if exist := userIDExists[task.Creator]; !exist {
			userIDs = append(userIDs, task.Creator)
			userIDExists[task.Creator] = true
		}
	}

	for idx, id := range taskIDs {
		if tasksInfoOfTask, ok := tasksInfo[id]; ok {
			inspectTotal := 0
			inspectCount := 0
			inspectNum := 0
			reviewTotal := 0
			reviewNum := 0
			task := &tasks[idx]

			staffs := map[string]bool{}
			for _, taskInfo := range *tasksInfoOfTask {
				if exist := userIDExists[taskInfo.StaffID]; !exist {
					userIDExists[taskInfo.StaffID] = true
					userIDs = append(userIDs, taskInfo.StaffID)
				}

				if taskInfo.Type == int8(0) {
					inspectTotal++
					if _, ok := staffs[taskInfo.StaffID]; !ok {
						inspectCount++
					}

					if taskInfo.Status == int8(1) {
						inspectNum++
					}
				} else {
					if taskInfo.Status == int8(1) {
						reviewNum++
					}
					task.Reviewer = taskInfo.StaffID
					reviewTotal++
				}

			}

			task.InspectTotal = inspectTotal
			task.InspectCount = inspectCount
			task.InspectNum = inspectNum
			task.ReviewTotal = reviewTotal
			task.ReviewNum = reviewNum
		}
	}

	authConn := authDB.Conn()
	usersMap, err := taskDao.Users(userIDs, authConn)
	if err != nil {
		return
	}

	for idx := range tasks {
		task := &tasks[idx]
		task.Creator = usersMap[task.Creator]

		if task.Reviewer != "" {
			task.Reviewer = usersMap[task.Reviewer]
		}
	}
	return
}
