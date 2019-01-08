package manual

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
	"net/http"
	"net/url"
	"strconv"
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
		Creator:    it.Creator,
		CreateTime: it.CreateTime,
		FormName:   it.Form.Name,
		TimeRange: CallTimeRange{
			StartTime: it.CallStart,
			EndTime:   it.CallEnd,
		},
		PublishTime:  it.PublishTime,
		InspectTotal: it.InspectTotal,
		InspectCount: it.InspectCount,
		InspectNum:   it.InspectNum,
		Reviewer:     it.Reviewer,
		ReviewNum:    it.ReviewNum,
		ReviewTotal:  it.ReviewTotal,
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
	values := r.URL.Query()
	filter := parseTaskFilter(&values)

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
