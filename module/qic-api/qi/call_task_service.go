package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	_ "emotibot.com/emotigo/pkg/logger"
)

func TasksByCalls(calls []model.Call) ([]*model.Task, error) {
	var err error
	callTasks := []*model.Task{}
	taskMap := map[int64]*model.Task{}
	for _, call := range calls {
		callTask, err := taskDao.CallTask(sqlConn, call)
		if err != nil {
			return callTasks, err
		}

		task, ok := taskMap[callTask.ID]
		if !ok {
			task = &callTask
			taskMap[callTask.ID] = task
			callTasks = append(callTasks, task)
		}

		if task.CallsOfStaffs == nil {
			task.CallsOfStaffs = map[string][]model.Call{}
		}

		_, ok = task.CallsOfStaffs[call.StaffID]
		if !ok {
			task.CallsOfStaffs[call.StaffID] = []model.Call{}
		}
		task.CallsOfStaffs[call.StaffID] = append(task.CallsOfStaffs[call.StaffID], call)
	}
	return callTasks, err
}
