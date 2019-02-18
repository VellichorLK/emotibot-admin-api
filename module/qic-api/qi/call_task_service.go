package qi

import (
	"fmt"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

// TaskByCall simply return the task that given call associated with.
// **Be ware: the CallsOfStaffs in returned task will only contains the given call.**
func TaskByCall(call model.Call) (*model.Task, error) {
	task, err := taskDao.CallTask(sqlConn, call)
	if err != nil {
		return nil, fmt.Errorf("get task by call '%d' failed, %v", call.ID, err)
	}

	task.CallsOfStaffs = map[string][]model.Call{}
	_, ok := task.CallsOfStaffs[call.StaffID]
	if !ok {
		task.CallsOfStaffs[call.StaffID] = []model.Call{}
	}
	task.CallsOfStaffs[call.StaffID] = append(task.CallsOfStaffs[call.StaffID], call)
	return &task, nil
}

// TasksByCalls fetch all unique tasks of the Call.
// If two call has the same Task, that task will has its CallOfStaffs by the call Staff ID.
// returned tasks have no order guarantee.
// **Be ware: the CallsOfStaffs in returned task will only contains the given calls.**
func TasksByCalls(calls []model.Call) ([]*model.Task, error) {
	uniqueTasks := map[int64][]int{}
	for idx, call := range calls {
		uniqueTasks[call.TaskID] = append(uniqueTasks[call.TaskID], idx)
	}
	tasks := []*model.Task{}
	for _, IdxGrp := range uniqueTasks {
		t, err := TaskByCall(calls[IdxGrp[0]])
		if err != nil {
			return nil, err
		}
		for _, idx := range IdxGrp {
			c := calls[idx]
			t.CallsOfStaffs[c.StaffID] = append(t.CallsOfStaffs[c.StaffID], c)
		}
	}
	return tasks, nil
}
