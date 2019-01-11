package manual

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/qi"
	"emotibot.com/emotigo/module/qic-api/util/general"
	_ "emotibot.com/emotigo/pkg/logger"
	"fmt"
	"math"
	"math/rand"
)

var (
	ErrNilTask = fmt.Errorf("Nil InspectorTask Error")
	ErrNoTask  = fmt.Errorf("The task does not exist")
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

	tasksInfo, err := taskDao.GetTasksInfoBy(taskFilter, manualConn)
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
						inspectNum++
					}

					if taskInfo.Status == int8(1) {
						inspectCount++
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
		user := usersMap[task.Creator]
		if user != nil {
			task.Creator = user.Name
		}

		if task.Reviewer != "" {
			reviewer := usersMap[task.Reviewer]
			if reviewer != nil {
				task.Reviewer = reviewer.Name
			}
		}
	}
	return
}

func copyTask(task *model.InspectTask) *model.InspectTask {
	outlines := make([]model.Outline, len(task.Outlines))
	for idx := range task.Outlines {
		outline := model.Outline{
			ID:   task.Outlines[idx].ID,
			Name: task.Outlines[idx].Name,
		}
		outlines[idx] = outline
	}

	return &model.InspectTask{
		ID:          task.ID,
		Name:        task.Name,
		Enterprise:  task.Enterprise,
		Description: task.Description,
		CallStart:   task.CallStart,
		CallEnd:     task.CallEnd,
		Status:      task.Status,
		CreateTime:  task.CreateTime,
		UpdateTime:  task.UpdateTime,
		Form: model.ScoreForm{
			ID:   task.Form.ID,
			Name: task.Form.Name,
		},
		PublishTime:       task.PublishTime,
		InspectNum:        task.InspectNum,
		InspectCount:      task.InspectCount,
		InspectTotal:      task.InspectTotal,
		InspectPercentage: task.InspectPercentage,
		InspectByPerson:   task.InspectByPerson,
		Reviewer:          task.Reviewer,
		ReviewNum:         task.ReviewNum,
		ReviewTotal:       task.ReviewTotal,
		ReviewPercentage:  task.ReviewPercentage,
		ReviewByPerson:    task.ReviewByPerson,
		ExcludeInspected:  task.ExcludeInspected,
		Type:              task.Type,
		Outlines:          outlines,
	}
}

func GetTasksOfUsers(filter *model.StaffTaskFilter) (total int64, fullTasks []model.InspectTask, err error) {
	manualConn := manualDB.Conn()

	total, err = taskDao.CountTaskInfoBy(filter, manualConn)
	if err != nil {
		return
	}

	taskInfos, err := taskDao.GetTasksInfoBy(filter, manualConn)
	if err != nil {
		return
	}

	taskIDs := []int64{}
	for k, _ := range taskInfos {
		taskIDs = append(taskIDs, k)
	}

	taskFilter := &model.InspectTaskFilter{
		ID: taskIDs,
	}

	tasks, err := taskDao.GetBy(taskFilter, manualConn)
	if err != nil {
		return
	}

	// reviewer id to readable name
	userIDs := []string{}
	for _, task := range tasks {
		if task.Reviewer != "" {
			userIDs = append(userIDs, task.Reviewer)
		}
	}

	authConn := authDB.Conn()
	usersMap, err := taskDao.Users(userIDs, authConn)
	if err != nil {
		return
	}

	fullTasks = []model.InspectTask{}
	for _, task := range tasks {
		infos, ok := taskInfos[task.ID]
		if ok {
			var cInspectTask *model.InspectTask
			var cReviewTask *model.InspectTask
			for _, info := range *infos {
				if info.Type == 0 {
					if cInspectTask == nil {
						cInspectTask = copyTask(&task)
					}
					cInspectTask.Type = info.Type
					cInspectTask.InspectTotal++

					if info.Status == 1 {
						cInspectTask.InspectCount++
					}
				} else {
					if cReviewTask == nil {
						cReviewTask = copyTask(&task)
					}
					cReviewTask.Type = info.Type
					cReviewTask.ReviewTotal++

					if info.Status == 1 {
						cReviewTask.ReviewNum++
					}

					user := usersMap[info.StaffID]
					cReviewTask.Reviewer = user.Name
				}
			}

			if cInspectTask != nil {
				fullTasks = append(fullTasks, *cInspectTask)
			}

			if cReviewTask != nil {
				fullTasks = append(fullTasks, *cReviewTask)
			}
		}
	}
	return
}

func GetUser(id string) (user *model.Staff, err error) {
	userIDs := []string{
		id,
	}

	authConn := authDB.Conn()

	usersMap, err := taskDao.Users(userIDs, authConn)
	if err != nil {
		return
	}

	user = usersMap[id]
	return
}

// UpdateTask only updates publish time & status
func UpdateTask(taskID int64, task *model.InspectTask) (err error) {
	tx, err := manualDB.Begin()
	if err != nil {
		return
	}
	defer manualDB.ClearTransition(tx)

	taskModified := &model.InspectTask{
		Status:           task.Status,
		PublishTime:      task.PublishTime,
		ExcludeInspected: -1,
	}

	err = taskDao.Update(taskID, taskModified, tx)
	if err != nil {
		return
	}
	manualDB.Commit(tx)
	return
}

func AssignInspectorTask(taskID int64, enterprise string, assignTask *AssignTask) (err error) {
	tx, err := manualDB.Begin()
	if err != nil {
		return
	}

	// get task
	filter := &model.InspectTaskFilter{
		ID: []int64{
			taskID,
		},
		Enterprise: enterprise,
	}
	tasks, err := taskDao.GetBy(filter, tx)
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		err = ErrNoTask
		return
	}

	task := tasks[0]

	// fetch matched calls
	callQuery := model.CallQuery{
		CallTimeStart: &task.CallStart,
		CallTimeEnd:   &task.CallEnd,
		StaffID:       assignTask.Users,
	}
	calls, err := qi.Calls(tx, callQuery)
	if err != nil {
		return
	}

	// calls to callTasks
	callTasks, err := qi.TasksByCalls(calls)
	if err != nil {
		return
	}

	// assign inspector to call task by sampling rules
	percentage := task.InspectPercentage
	taskType := int8(0)
	// byperson := task.InspectByPerson
	if assignTask.Type == "reviewer" {
		percentage = assignTask.Sampling.Percentage
		taskType = int8(1)
		// byperson = assignTask.Sampling.ByPerson
	}
	// byperson := task.InspectByPerson
	toInspectTasks := []*model.Task{}

	if percentage != 0 {
		total := len(callTasks)
		totalInspect := math.Ceil(float64(total*percentage) / float64(100))
		end := int64(totalInspect)
		toInspectTasks = callTasks[0:end]
	} else {
		// TODO: by person need to check
		toInspectTasks = callTasks
	}

	// inspector to call task
	assigns := []model.StaffTaskInfo{}
	inspectorNum := len(assignTask.Users)
	for _, task := range toInspectTasks {
		idx := rand.Intn(inspectorNum)
		inspectorID := assignTask.Users[idx]

		assignTask := model.StaffTaskInfo{
			StaffID: inspectorID,
			TaskID:  taskID,
			CallID:  task.ID,
			Type:    taskType,
		}
		assigns = append(assigns, assignTask)
	}

	err = taskDao.AssignInspectTasks(assigns, tx)
	if err != nil {
		return
	}

	err = manualDB.Commit(tx)
	return
}

func GetCallsOfUser(userID string) (calls []model.Call, err error) {
	return
}

func GetOutlines() ([]*model.Outline, error) {
	manualConn := manualDB.Conn()
	return taskDao.Outlines(manualConn)
}

func GetUsers(userType string) ([]*model.Staff, error) {
	authConn := authDB.Conn()
	return taskDao.UsersByType(userType, authConn)
}

func FinishTask(staff string, callID int64) error {
	manualConn := manualDB.Conn()
	return taskDao.FinishTask(staff, callID, manualConn)
}

func GetScoreForms() ([]*model.ScoreForm, error) {
	manualConn := manualDB.Conn()
	return taskDao.ScoreForms(manualConn)
}
