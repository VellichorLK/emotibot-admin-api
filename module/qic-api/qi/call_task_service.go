package qi

import (
	"fmt"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var (
	callTask = taskDao.CallTask
	tasks    = taskDao.Task
)

var (
	taskByCall = func(call model.Call) (*model.Task, error) {
		tasks, err := TasksByCalls([]model.Call{call})
		if err != nil {
			return nil, err
		}
		if len(tasks) != 1 {
			return nil, fmt.Errorf("call task id '%d' does not exist", call.TaskID)
		}
		return tasks[0], nil
	}
	tasksByCalls = func(calls []model.Call) ([]*model.Task, error) {
		// indexes is a join of given calls' unique task.
		indexes := map[int64][]int{}
		for idx, call := range calls {
			indexes[call.TaskID] = append(indexes[call.TaskID], idx)
		}
		taskIDs := make([]int64, 0)
		for taskID := range indexes {
			taskIDs = append(taskIDs, taskID)
		}
		// all tasks of given calls
		uniqueTasks, err := tasks(nil, model.TaskQuery{ID: taskIDs})
		if err != nil {
			return nil, fmt.Errorf("get tasks failed, %v", err)
		}
		data := []*model.Task{}
		for i := 0; i < len(uniqueTasks); i++ {
			t := &uniqueTasks[i]
			callIndexes, exist := indexes[t.ID]
			if !exist {
				return nil, fmt.Errorf("data inconsistent, task id '%d' does not exist", t.ID)
			}
			t.CallsOfStaffs = make(map[string][]model.Call)
			for _, idx := range callIndexes {
				c := calls[idx]
				t.CallsOfStaffs[c.StaffID] = append(t.CallsOfStaffs[c.StaffID], c)
			}
			data = append(data, t)
		}
		return data, nil
	}
)

// TaskByCall simply return the task that given call associated with.
// **Be ware: the CallsOfStaffs in returned task will only contains the given call.**
func TaskByCall(call model.Call) (*model.Task, error) {
	return taskByCall(call)
}

// TasksByCalls fetch all unique tasks of the Call.
// If two call has the same Task, that task will has its CallOfStaffs by the call Staff ID.
// returned tasks are ordered by tasks ID.
// **Be ware: the CallsOfStaffs in returned task will only contains the given calls.**
func TasksByCalls(calls []model.Call) ([]*model.Task, error) {
	return tasksByCalls(calls)
}
