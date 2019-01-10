package manual

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
)

const (
	ADMIN_USER  = 1
	NORMAL_USER = 2
)

type CallTimeRange struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type Sampling struct {
	Percentage int `json:"percentage"`
	ByPerson   int `json:"byperson"`
}

type InspectTaskInReq struct {
	Name         string        `json:"task_name"`
	TimeRange    CallTimeRange `json:"call_time_range"`
	Outlines     []int64       `json:"outline_ids"`
	Staffs       []string      `json:"staff_ids"`
	Form         int64         `json:"form_id"`
	Sampling     Sampling      `json:"sampling_rule"`
	IsInspecting int8          `json:"is_inspecting"`
	PublishTime  int64         `json:"publish_time"`
}

type InspectTaskInRes struct {
	ID           int64         `json:"task_id"`
	Name         string        `json:"task_name"`
	TimeRange    CallTimeRange `json:"call_time_range"`
	Status       int8          `json:"task_status"`
	Creator      string        `json:"task_creator"`
	CreateTime   int64         `json:"create_time"`
	FormName     string        `json:"form_name"`
	Outlines     []string      `json:"outlines"`
	PublishTime  int64         `json:"publish_time"`
	InspectNum   int           `json:"inspector_num"`
	InspectCount int           `json:"inspect_count"`
	InspectTotal int           `json:"inspect_total"`
	Reviewer     string        `json:"reviewer"`
	ReviewNum    int           `json:"review_count"`
	ReviewTotal  int           `json:"review_total"`
}

type InspectTaskInResFromNormalUser struct {
	ID          int64         `json:"task_id"`
	Name        string        `json:"task_name"`
	TimeRange   CallTimeRange `json:"call_time_range"`
	Status      int8          `json:"task_status"`
	Creator     string        `json:"task_creator"`
	CreateTime  int64         `json:"create_time"`
	FormName    string        `json:"form_name"`
	Outlines    []string      `json:"outlines"`
	Type        int8          `json:"task_type"`
	Count       int           `json:"count"`
	Total       int           `json:"total"`
	Reviewer    string        `json:"reviewer"`
	PublishTime int64         `json:"publish_time"`
}

type AssignTask struct {
	Type     string   `json:"assign_type"`
	Users    []string `json:"user_ids"`
	Sampling Sampling `json:"sampling_rule"`
}

func inspectTaskInReqToInspectTask(inreq *InspectTaskInReq) (task *model.InspectTask) {
	taskOutlines := make([]model.Outline, len(inreq.Outlines))
	for idx := range inreq.Outlines {
		taskOutlines[idx] = model.Outline{
			ID: inreq.Outlines[idx],
		}
	}

	task = &model.InspectTask{
		Name:             inreq.Name,
		Outlines:         taskOutlines,
		Staffs:           inreq.Staffs,
		ExcludeInspected: inreq.IsInspecting,
		Form: model.ScoreForm{
			ID: inreq.Form,
		},
		InspectPercentage: inreq.Sampling.Percentage,
		InspectByPerson:   inreq.Sampling.ByPerson,
		CallStart:         inreq.TimeRange.StartTime,
		CallEnd:           inreq.TimeRange.EndTime,
		PublishTime:       inreq.PublishTime,
	}
	return
}

func inspectTaskToInspectTaskInRes(it *model.InspectTask) *InspectTaskInRes {
	outlines := make([]string, len(it.Outlines))
	for idx, outline := range it.Outlines {
		outlines[idx] = outline.Name
	}

	inRes := &InspectTaskInRes{
		ID:         it.ID,
		Name:       it.Name,
		Outlines:   outlines,
		Status:     it.Status,
		CreateTime: it.CreateTime,
		FormName:   it.Form.Name,
		TimeRange: CallTimeRange{
			StartTime: it.CallStart,
			EndTime:   it.CallEnd,
		},
		PublishTime:  it.PublishTime,
		Reviewer:     it.Reviewer,
		InspectNum:   it.InspectNum,
		InspectCount: it.InspectCount,
		InspectTotal: it.InspectTotal,
		ReviewNum:    it.ReviewNum,
		ReviewTotal:  it.ReviewTotal,
	}
	return inRes
}

func inspectTaskToInspectTaskInResForNormalUser(it *model.InspectTask) *InspectTaskInResFromNormalUser {
	outlines := make([]string, len(it.Outlines))
	for idx, outline := range it.Outlines {
		outlines[idx] = outline.Name
	}

	inRes := &InspectTaskInResFromNormalUser{
		ID:         it.ID,
		Name:       it.Name,
		Outlines:   outlines,
		Status:     it.Status,
		Creator:    it.Creator,
		CreateTime: it.CreateTime,
		FormName:   it.Form.Name,
		TimeRange: CallTimeRange{
			StartTime: it.CallStart,
			EndTime:   it.CallEnd,
		},
		PublishTime: it.PublishTime,
		Reviewer:    it.Reviewer,
	}

	if it.Type == int8(0) {
		inRes.Total = it.InspectTotal
		inRes.Count = it.InspectCount
	} else if it.Type == int8(1) {
		inRes.Total = it.ReviewTotal
		inRes.Count = it.ReviewNum
	}
	return inRes
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	inreq := InspectTaskInReq{}
	err := util.ReadJSON(r, &inreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := inspectTaskInReqToInspectTask(&inreq)
	task.Enterprise = requestheader.GetEnterpriseID(r)
	task.Creator = requestheader.GetUserID(r)

	id, err := CreateTask(task)
	if err != nil {
		logger.Error.Printf("error while create inspect task in handleCreateTask, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		ID int64 `json:"task_id"`
	}

	response := Response{
		ID: id,
	}

	util.WriteJSON(w, response)
}

func parseTaskFilter(values *url.Values) *model.InspectTaskFilter {
	filter := &model.InspectTaskFilter{}
	var err error

	pageStr := values.Get("page")
	filter.Page, err = strconv.Atoi(pageStr)
	if err != nil {
		filter.Page = 0
	}

	limitStr := values.Get("limit")
	filter.Limit, err = strconv.Atoi(limitStr)
	if err != nil {
		filter.Limit = 10
	}
	return filter
}

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	userID := requestheader.GetUserID(r)
	user, err := GetUser(userID)
	if err != nil {
		logger.Error.Printf("error while get user in handleGetTasks, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	values := r.URL.Query()
	filter := parseTaskFilter(&values)

	if user.Type == ADMIN_USER {
		handleAdminUserGetTasks(filter, w)
	} else if user.Type == NORMAL_USER {
		handleNormalUserGetTasks(userID, filter, w)
	}
}

func handleAdminUserGetTasks(filter *model.InspectTaskFilter, w http.ResponseWriter) {
	total, tasks, err := GetTasks(filter)
	if err != nil {
		logger.Error.Printf("error while get inspect tasks in handleGetTasks, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Paging general.Paging      `json:"paging"`
		Data   []*InspectTaskInRes `json:"data"`
	}

	tasksInRes := make([]*InspectTaskInRes, len(tasks))
	for idx, task := range tasks {
		tasksInRes[idx] = inspectTaskToInspectTaskInRes(&task)
	}

	response := Response{
		Paging: general.Paging{
			Total: total,
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Data: tasksInRes,
	}

	util.WriteJSON(w, response)
}

func handleNormalUserGetTasks(userID string, filter *model.InspectTaskFilter, w http.ResponseWriter) {
	taskFilter := &model.StaffTaskFilter{
		Page:  filter.Page,
		Limit: filter.Limit,
		StaffIDs: []string{
			userID,
		},
	}
	total, tasks, err := GetTasksOfUsers(taskFilter)
	if err != nil {
		logger.Error.Printf("error while get inspect tasks in handleGetTasks, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Paging general.Paging                    `json:"paging"`
		Data   []*InspectTaskInResFromNormalUser `json:"data"`
	}

	tasksInRes := make([]*InspectTaskInResFromNormalUser, len(tasks))
	for idx, task := range tasks {
		tasksInRes[idx] = inspectTaskToInspectTaskInResForNormalUser(&task)
	}

	response := Response{
		Paging: general.Paging{
			Total: total,
			Page:  filter.Page,
			Limit: filter.Limit,
		},
		Data: tasksInRes,
	}

	util.WriteJSON(w, response)
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	userID := requestheader.GetUserID(r)
	user, err := GetUser(userID)
	if err != nil {
		logger.Error.Printf("error while get user in handleGetTasks, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if user.Type == int8(1) {
		handleAdminUserGetTask(w, r)
	} else if user.Type == int8(2) {
		handleNormalUserGetTask(w, r)
	}
}

func handleAdminUserGetTask(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	taskIDstr := general.ParseID(r)
	taskID, err := strconv.ParseInt(taskIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := &model.InspectTaskFilter{
		ID: []int64{
			taskID,
		},
		Enterprise: enterprise,
	}

	_, tasks, err := GetTasks(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tasks) == 0 {
		http.NotFound(w, r)
		return
	}

	task := tasks[0]
	taskInRes := inspectTaskToInspectTaskInRes(&task)

	util.WriteJSON(w, taskInRes)
}

func handleNormalUserGetTask(w http.ResponseWriter, r *http.Request) {
	userID := requestheader.GetUserID(r)

	taskIDstr := general.ParseID(r)
	taskID, err := strconv.ParseInt(taskIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskFilter := &model.StaffTaskFilter{
		StaffIDs: []string{
			userID,
		},
		TaskIDs: []int64{
			taskID,
		},
	}

	_, tasks, err := GetTasksOfUsers(taskFilter)
	if err != nil {
		logger.Error.Printf("error while get inspect tasks in handleGetTasks, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tasks) == 0 {
		http.NotFound(w, r)
		return
	}

	task := tasks[0]
	taskInRes := inspectTaskToInspectTaskInResForNormalUser(&task)

	util.WriteJSON(w, taskInRes)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	userID := requestheader.GetUserID(r)
	user, err := GetUser(userID)
	if err != nil {
		logger.Error.Printf("error while get user in handleUpdateTask, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil || user.Type == NORMAL_USER {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	taskIDstr := general.ParseID(r)
	taskID, err := strconv.ParseInt(taskIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inreq := InspectTaskInReq{}
	err = util.ReadJSON(r, &inreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Info.Printf("inreq: %+v", inreq)

	task := inspectTaskInReqToInspectTask(&inreq)
	task.Status = int8(-1)

	err = UpdateTask(taskID, task)
	if err != nil {
		logger.Error.Printf("error while update inspect task in handleUpdateTask, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleAssignStaffToTask(w http.ResponseWriter, r *http.Request) {
	taskIDstr := general.ParseID(r)
	taskID, err := strconv.ParseInt(taskIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	enterprise := requestheader.GetEnterpriseID(r)

	userID := requestheader.GetUserID(r)
	user, err := GetUser(userID)
	if err != nil {
		logger.Error.Printf("error while get user in handleAssignStaffToTask, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil || user.Type == NORMAL_USER {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	assigns := AssignTask{}
	err = util.ReadJSON(r, &assigns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = AssignInspectorTask(taskID, enterprise, &assigns)
	if err != nil {
		logger.Error.Printf("error while assign tasks in handleAssignStaffToTask, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func handleInspectTaskPublish(w http.ResponseWriter, r *http.Request) {
	userID := requestheader.GetUserID(r)
	user, err := GetUser(userID)
	if err != nil {
		logger.Error.Printf("error while get user in handleInspectTaskPublish, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil || user.Type == NORMAL_USER {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	taskIDstr := general.ParseID(r)
	taskID, err := strconv.ParseInt(taskIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	published := vars["published"]

	task := &model.InspectTask{}
	if published == "publish" {
		task.Status = 1
	} else {
		task.Status = 0
	}

	err = UpdateTask(taskID, task)
	if err != nil {
		logger.Error.Printf("error while update task status in handleInspectTaskPublish, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleUserAssignedCalls(w http.ResponseWriter, r *http.Request) {
	// userID := requestheader.GetUserID(r)

	// GetTasksOfUsers()
}

func handleGetOutlines(w http.ResponseWriter, r *http.Request) {
	outlines, err := GetOutlines()
	if err != nil {
		logger.Error.Printf("error while get outlines in handleGetOutline, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, outlines)
}

func handleGetInspectors(w http.ResponseWriter, r *http.Request) {
	inspectors, err := GetUsers("inspector")

	if err != nil {
		logger.Error.Printf("error while get outlines in handleGetOutline, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, inspectors)
}

func handleGetCustomerStaffs(w http.ResponseWriter, r *http.Request) {
	staffs, err := GetUsers("staff")

	if err != nil {
		logger.Error.Printf("error while get outlines in handleGetOutline, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, staffs)
}

func handleFinishInspectTask(w http.ResponseWriter, r *http.Request) {
	userID := requestheader.GetUserID(r)

	callIDstr := general.ParseID(r)
	callID, err := strconv.ParseInt(callIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = FinishTask(userID, callID)
	if err != nil {
		logger.Error.Printf("error while set inspect task finished in handleFinishInspectTask, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
