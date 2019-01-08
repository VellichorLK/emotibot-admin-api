package model

import (
	_ "emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
)

type InspectTaskFilter struct {
	UUID       []string
	Page       int
	Limit      int
	Enterprise string
}

type InspectTask struct {
	ID                int64
	UUID              string
	Name              string
	Enterprise        string
	Description       string
	CallStart         int64
	CallEnd           int64
	Status            int8
	Creator           string
	CreateTime        int64
	UpdateTime        int64
	Form              ScoreForm
	Outlines          []Outline
	PublishTime       int64
	InspectNum        int
	InspectCount      int
	InspectTotal      int
	InspectPercentage int
	InspectByPerson   int
	Reviewer          string
	ReviewNum         int
	ReviewTotal       int
	ReviewPercentage  int
	ReviewByPerson    int
	Staffs            []string
	ExcludeInspected  int8
}

type StaffTaskInfo struct {
	TaskID    int64
	StaffID   string
	StaffName string
	CallID    int64
	Status    int8
	Type      int8
}

type StaffTaskFilter struct {
	TaskIDs  []int64
	CallIDs  []int64
	StaffIDs []string
}

type Outline struct {
	ID   int64
	Name string
}

type ScoreForm struct {
	ID   int64
	Name string
}

type InspectTaskDao interface {
	Create(task *InspectTask, sql SqlLike) (int64, error)
	CountBy(filter *InspectTaskFilter, sql SqlLike) (int64, error)
	GetBy(filter *InspectTaskFilter, sql SqlLike) ([]InspectTask, error)
	Users(uids []string, sql SqlLike) (map[string]string, error)
	TasksInfoBy(filter *StaffTaskFilter, sql SqlLike) (map[int64]*[]StaffTaskInfo, error)
}

type InspectTaskSqlDao struct{}

func (dao *InspectTaskSqlDao) Create(task *InspectTask, sql SqlLike) (id int64, err error) {
	if task == nil {
		err = fmt.Errorf("Nil Task Error")
		return
	}

	fields := []string{
		fldName,
		fldEnterprise,
		fldDescription,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
		ITCallStart,
		ITCallEnd,
		ITInspectPercentage,
		ITInspectByPerson,
		fldCreator,
		ITExcluedInspected,
		ITFormID,
	}
	values := []interface{}{
		task.Name,
		task.Enterprise,
		task.Description,
		task.CreateTime,
		task.UpdateTime,
		task.UUID,
		task.CallStart,
		task.CallEnd,
		task.InspectPercentage,
		task.InspectByPerson,
		task.Creator,
		task.ExcludeInspected,
		task.Form.ID,
	}

	insertStr := fmt.Sprintf(
		"INSERT INTO `%s` (`%s`) VALUES (%s)",
		tblInspectTask,
		strings.Join(fields, "`, `"),
		fmt.Sprintf("?%s", strings.Repeat(", ?", len(values)-1)),
	)

	result, err := sql.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while create inspect task in dao.Create, err: %s", err.Error())
		return
	}

	id, err = result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get inspect task id in dao.Create, err: %s", err.Error())
		return
	}

	outlineNum := len(task.Outlines)
	if outlineNum > 0 {
		fields = []string{
			RITOTaskID,
			RITOTOutlineID,
		}

		valueStr := fmt.Sprintf("(?, ?)%s", strings.Repeat(", (?, ?)", outlineNum-1))
		values = make([]interface{}, 0)
		for _, outline := range task.Outlines {
			values = append(values, id, outline.ID)
		}

		insertStr = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s",
			tblRelITOutline,
			strings.Join(fields, ","),
			valueStr,
		)

		_, err = sql.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert outline relation in dao.Create, err: %s", err.Error())
			return
		}
	}
	return
}

func queryInspectTaskSQLBy(filter *InspectTaskFilter) (queryStr string, values []interface{}) {
	values = []interface{}{}
	conditionStr := "WHERE "
	conditions := []string{}

	if filter.Enterprise != "" {
		conditions = append(conditions, fmt.Sprintf("%s=?", fldEnterprise))
		values = append(values, filter.Enterprise)
	}

	if len(filter.UUID) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.UUID)-1))
		conditions = append(conditions, fmt.Sprintf("%s IN(%s)", fldUUID, idStr))

		for _, uuid := range filter.UUID {
			values = append(values, uuid)
		}
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	} else {
		conditionStr = ""
	}

	queryStr = fmt.Sprintf(
		`SELECT it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, 
		form.%s as fname, ot.%s as otname FROM (SELECT * FROM %s %s) as it
		LEFT JOIN %s as form ON it.%s = form.%s
		LEFT JOIN %s as ritol ON it.%s = ritol.%s
		LEFT JOIN %s as ot ON ritol.%s = ot.%s`,
		fldID,
		fldName,
		ITCallStart,
		ITCallEnd,
		fldStatus,
		fldCreator,
		fldCreateTime,
		ITPublishTime,
		fldName,
		fldName,
		tblInspectTask,
		conditionStr,
		tblScoreForm,
		ITFormID,
		fldID,
		tblRelITOutline,
		fldID,
		RITOTaskID,
		tblOutline,
		RITOTOutlineID,
		fldID,
	)
	return
}

func (dao *InspectTaskSqlDao) CountBy(filter *InspectTaskFilter, sql SqlLike) (total int64, err error) {
	queryStr, values := queryInspectTaskSQLBy(filter)
	queryStr = fmt.Sprintf(
		"SELECT COUNT(it.%s) FROM (SELECT it.%s FROM (SELECT it.%s FROM (%s) as it) as it GROUP BY %s) as it",
		fldID,
		fldID,
		fldID,
		queryStr,
		fldID,
	)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while count inspect tasks in dao.CountBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (dao *InspectTaskSqlDao) GetBy(filter *InspectTaskFilter, sql SqlLike) (tasks []InspectTask, err error) {
	queryStr, values := queryInspectTaskSQLBy(filter)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get inspect tasks in dao.CountBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	tasks = []InspectTask{}
	var cTask *InspectTask
	for rows.Next() {
		form := ScoreForm{}
		outline := Outline{}
		task := InspectTask{}
		rows.Scan(
			&task.ID,
			&task.Name,
			&task.CallStart,
			&task.CallEnd,
			&task.Status,
			&task.Creator,
			&task.CreateTime,
			&task.PublishTime,
			&form.Name,
			&outline.Name,
		)

		if cTask == nil || cTask.ID != task.ID {
			if cTask != nil {
				tasks = append(tasks, *cTask)
			}
			cTask = &task
		}
		cTask.Outlines = append(cTask.Outlines, outline)
	}

	if cTask != nil {
		tasks = append(tasks, *cTask)
	}
	return
}

func (dao *InspectTaskSqlDao) Users(uids []string, sql SqlLike) (users map[string]string, err error) {
	idCondition := ""
	values := make([]interface{}, len(uids))
	if len(uids) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(uids)-1))
		idCondition = fmt.Sprintf("WHERE %s IN (%s)", fldUUID, idStr)

		for idx, uid := range uids {
			values[idx] = uid
		}
	}

	queryStr := fmt.Sprintf(
		"SELECT %s, %s FROM %s %s",
		fldUUID,
		USERDisplayName,
		tblUsers,
		idCondition,
	)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get users in dao.Users, err: %s", err.Error())
		return
	}
	defer rows.Close()

	users = map[string]string{}
	for rows.Next() {
		uuid := ""
		userName := ""
		rows.Scan(
			&uuid,
			&userName,
		)
		users[uuid] = userName
	}
	return
}

func (dao *InspectTaskSqlDao) TasksInfoBy(filter *StaffTaskFilter, sql SqlLike) (taskInfos map[int64]*[]StaffTaskInfo, err error) {
	conditionStr := ""
	conditions := []string{}
	values := []interface{}{}
	if len(filter.TaskIDs) > 0 {
		idStr := "?"
		idStr = fmt.Sprintf("%s %s", idStr, strings.Repeat(", ?", len(filter.TaskIDs)-1))
		for _, id := range filter.TaskIDs {
			values = append(values, id)
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", RITCSTaskID, idStr))
	}

	if len(filter.CallIDs) > 0 {
		idStr := fmt.Sprintf("?", filter.CallIDs[0])
		idStr = fmt.Sprintf("%s %s", idStr, strings.Repeat(", ?", len(filter.CallIDs)-1))
		for _, id := range filter.CallIDs {
			values = append(values, id)
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", RITCSCallID, idStr))
	}

	if len(filter.StaffIDs) > 0 {
		idStr := fmt.Sprintf("?", filter.StaffIDs[0])
		idStr = fmt.Sprintf("%s %s", idStr, strings.Repeat(", ?", len(filter.StaffIDs)-1))
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", RITCSStaffID, idStr))
		for _, id := range filter.StaffIDs {
			values = append(values, id)
		}
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("WHERE %s", strings.Join(conditions, " and "))
	}

	queryStr := fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s FROM %s %s",
		RITCSTaskID,
		RITCSCallID,
		RITCSStaffID,
		fldStatus,
		fldType,
		tblRelITCallStaff,
		conditionStr,
	)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while query task info in dao.TasksInfoBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	taskInfos = map[int64]*[]StaffTaskInfo{}
	for rows.Next() {
		taskInfo := StaffTaskInfo{}
		rows.Scan(
			&taskInfo.TaskID,
			&taskInfo.CallID,
			&taskInfo.StaffID,
			&taskInfo.Status,
			&taskInfo.Type,
		)

		if taskInfosOfTask, ok := taskInfos[taskInfo.TaskID]; !ok {
			infos := []StaffTaskInfo{
				taskInfo,
			}
			taskInfos[taskInfo.TaskID] = &infos
		} else {
			infos := append(*taskInfosOfTask, taskInfo)
			taskInfosOfTask = &infos
		}
	}
	return
}
