package model

import (
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
)

type InspectTaskFilter struct {
	ID         []int64
	UUID       []string
	Page       int
	Limit      int
	Enterprise string
	Outlines   []int64
	Staffs     []string
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
	Type              int8
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
	Page     int
	Limit    int
}

type Outline struct {
	ID   int64  `json:"id"`
	Name string `json:"text"`
}

type ScoreForm struct {
	ID   int64  `json:"form_id"`
	Name string `json:"form_name"`
}

type Staff struct {
	UUID string `json:"id"`
	Name string `json:"name"`
	Type int8   `json:"-"`
}

type InspectTaskDao interface {
	Create(task *InspectTask, sql SqlLike) (int64, error)
	CountBy(filter *InspectTaskFilter, sql SqlLike) (int64, error)
	GetBy(filter *InspectTaskFilter, sql SqlLike) ([]InspectTask, error)
	Users(uids []string, sql SqlLike) (map[string]*Staff, error)
	CountTaskInfoBy(filter *StaffTaskFilter, sql SqlLike) (int64, error)
	GetTasksInfoBy(filter *StaffTaskFilter, sql SqlLike) (map[int64]*[]StaffTaskInfo, error)
	Update(taskID int64, task *InspectTask, sql SqlLike) error
	AssignInspectTasks(assigns []StaffTaskInfo, sql SqlLike) error
	Outlines(SqlLike) ([]*Outline, error)
	UsersByType(string, SqlLike) ([]*Staff, error)
	FinishTask(string, int64, SqlLike) error
	ScoreForms(SqlLike) ([]*ScoreForm, error)
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

	staffNum := len(task.Staffs)
	if staffNum != 0 {
		fields = []string{
			RITStaffTaskID,
			RITStaffStaffID,
		}

		valuesStr := fmt.Sprintf("(?, ?)%s", strings.Repeat(", (?, ?)", staffNum-1))
		values = []interface{}{}
		for _, staff := range task.Staffs {
			values = append(values, id, staff)
		}

		insertStr = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s",
			tblRelITStaff,
			strings.Join(fields, ","),
			valuesStr,
		)

		_, err = sql.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert staff relation in dao.Create, err: %s", err.Error())
			return
		}
	}
	return
}

func queryInspectTaskSQLBy(filter *InspectTaskFilter, doPage bool) (queryStr string, values []interface{}) {
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

	doPageStr := ""
	if doPage && filter.Limit > 0 {
		start := filter.Page * filter.Limit
		end := start + filter.Limit
		doPageStr = fmt.Sprintf("LIMIT %d, %d", start, end)
	}

	outlineCondition := fmt.Sprintf("LEFT JOIN %s", tblRelITOutline)
	outlineNum := len(filter.Outlines)
	if outlineNum > 0 {
		outlineCondition = fmt.Sprintf(
			"INNER JOIN (SELECT * FROM %s WHERE %s IN (%s))",
			tblRelITStaff,
			RITOTOutlineID,
			fmt.Sprintf("?%s", strings.Repeat(", ?", outlineNum-1)),
		)

		for _, outlineID := range filter.Outlines {
			values = append(values, outlineID)
		}
	}

	staffCondition := fmt.Sprintf("LEFT JOIN %s", tblRelITStaff)
	staffNum := len(filter.Staffs)
	if staffNum > 0 {
		staffCondition = fmt.Sprintf(
			"INNER JOIN (SELECT * FROM %s WHERE %s IN (%s))",
			tblRelITStaff,
			RITStaffStaffID,
			fmt.Sprintf("?%s", strings.Repeat(", ?", staffNum-1)),
		)

		for _, staffID := range filter.Staffs {
			values = append(values, staffID)
		}
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	} else {
		conditionStr = ""
	}

	queryStr = fmt.Sprintf(
		`SELECT it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, it.%s, it.%s,
		form.%s as fname, ot.%s as otname, rits.%s as staff_id FROM (SELECT * FROM %s %s %s) as it
		LEFT JOIN %s as form ON it.%s = form.%s
		%s as ritol ON it.%s = ritol.%s
		LEFT JOIN %s as ot ON ritol.%s = ot.%s
		%s as rits ON rits.%s = it.%s`,
		fldID,
		fldName,
		ITCallStart,
		ITCallEnd,
		fldStatus,
		fldCreator,
		fldCreateTime,
		ITPublishTime,
		ITInspectPercentage,
		ITInspectByPerson,
		fldName,
		fldName,
		RITStaffStaffID,
		tblInspectTask,
		conditionStr,
		doPageStr,
		tblScoreForm,
		ITFormID,
		fldID,
		outlineCondition,
		fldID,
		RITOTaskID,
		tblOutline,
		RITOTOutlineID,
		fldID,
		staffCondition,
		RITStaffTaskID,
		fldID,
	)
	logger.Info.Printf("queryStr: %s\n", queryStr)
	return
}

func (dao *InspectTaskSqlDao) CountBy(filter *InspectTaskFilter, sql SqlLike) (total int64, err error) {
	queryStr, values := queryInspectTaskSQLBy(filter, false)
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
	queryStr, values := queryInspectTaskSQLBy(filter, true)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get inspect tasks in dao.GetBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	tasks = []InspectTask{}
	var cTask *InspectTask
	staffMap := map[string]bool{}
	outlineMap := map[string]bool{}
	for rows.Next() {
		form := ScoreForm{}
		outline := Outline{}
		task := InspectTask{}
		staff := ""
		rows.Scan(
			&task.ID,
			&task.Name,
			&task.CallStart,
			&task.CallEnd,
			&task.Status,
			&task.Creator,
			&task.CreateTime,
			&task.PublishTime,
			&task.InspectPercentage,
			&task.InspectByPerson,
			&form.Name,
			&outline.Name,
			&staff,
		)

		if cTask == nil || cTask.ID != task.ID {
			if cTask != nil {
				tasks = append(tasks, *cTask)
			}
			cTask = &task

			staffMap = map[string]bool{}
			outlineMap = map[string]bool{}
		}

		if _, ok := outlineMap[outline.Name]; !ok {
			cTask.Outlines = append(cTask.Outlines, outline)
			outlineMap[outline.Name] = true
		}

		if _, ok := staffMap[staff]; !ok {
			cTask.Staffs = append(cTask.Staffs, staff)
			staffMap[staff] = true
		}
	}

	if cTask != nil {
		tasks = append(tasks, *cTask)
	}
	return
}

func (dao *InspectTaskSqlDao) Users(uids []string, sql SqlLike) (users map[string]*Staff, err error) {
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
		"SELECT %s, %s, %s FROM %s %s",
		fldUUID,
		USERDisplayName,
		fldType,
		tblUsers,
		idCondition,
	)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get users in dao.Users, err: %s", err.Error())
		return
	}
	defer rows.Close()

	users = map[string]*Staff{}
	for rows.Next() {
		user := Staff{}

		rows.Scan(
			&user.UUID,
			&user.Name,
			&user.Type,
		)
		users[user.UUID] = &user
	}
	return
}

func queryInspectTaskInfoSQLBy(filter *StaffTaskFilter) (queryStr string, values []interface{}) {
	conditionStr := ""
	conditions := []string{}
	values = []interface{}{}

	if len(filter.TaskIDs) > 0 {
		idStr := "?"
		idStr = fmt.Sprintf("%s %s", idStr, strings.Repeat(", ?", len(filter.TaskIDs)-1))
		for _, id := range filter.TaskIDs {
			values = append(values, id)
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", RITCSTaskID, idStr))
	}

	if len(filter.CallIDs) > 0 {
		idStr := "?"
		idStr = fmt.Sprintf("%s %s", idStr, strings.Repeat(", ?", len(filter.CallIDs)-1))
		for _, id := range filter.CallIDs {
			values = append(values, id)
		}
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", RITCSCallID, idStr))
	}

	if len(filter.StaffIDs) > 0 {
		idStr := "?"
		idStr = fmt.Sprintf("%s %s", idStr, strings.Repeat(", ?", len(filter.StaffIDs)-1))
		conditions = append(conditions, fmt.Sprintf("%s IN (%s)", RITCSStaffID, idStr))
		for _, id := range filter.StaffIDs {
			values = append(values, id)
		}
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("WHERE %s", strings.Join(conditions, " and "))
	}

	queryStr = fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s FROM %s %s",
		RITCSTaskID,
		RITCSCallID,
		RITCSStaffID,
		fldStatus,
		fldType,
		tblRelITCallStaff,
		conditionStr,
	)
	return
}

func (dao *InspectTaskSqlDao) CountTaskInfoBy(filter *StaffTaskFilter, sql SqlLike) (total int64, err error) {
	queryStr, values := queryInspectTaskInfoSQLBy(filter)

	queryStr = fmt.Sprintf(
		"SELECT COUNT(info.%s) FROM (SELECT info.%s FROM (SELECT info.%s FROM (%s) as info) as info GROUP BY %s) as info",
		RITCSTaskID,
		RITCSTaskID,
		RITCSTaskID,
		queryStr,
		RITCSTaskID,
	)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while count task information in dao.CountTaskInfoBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (dao *InspectTaskSqlDao) GetTasksInfoBy(filter *StaffTaskFilter, sql SqlLike) (taskInfos map[int64]*[]StaffTaskInfo, err error) {
	queryStr, values := queryInspectTaskInfoSQLBy(filter)

	if filter.Limit > 0 {
		start := filter.Page * filter.Limit
		end := start + filter.Limit
		queryStr = fmt.Sprintf("%s Limit %d, %d", queryStr, start, end)
	}

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

func (dao *InspectTaskSqlDao) Update(taskID int64, task *InspectTask, sql SqlLike) (err error) {
	updateFields := []string{}
	values := []interface{}{}
	if task.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", fldName))
		values = append(values, task.Name)
	}

	if task.CallStart > 0 {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", ITCallStart))
		values = append(values, task.CallStart)
	}

	if task.CallEnd > 0 {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", ITCallEnd))
		values = append(values, task.CallEnd)
	}

	if task.Description != "" {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", fldDescription))
		values = append(values, task.Description)
	}

	if task.ExcludeInspected > -1 {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", ITExcluedInspected))
		values = append(values, task.ExcludeInspected)
	}

	if task.Form.ID > 0 {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", ITFormID))
		values = append(values, task.Form.ID)
	}

	if task.Status > -1 {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", fldStatus))
		values = append(values, task.Status)
	}

	if task.PublishTime > 0 {
		updateFields = append(updateFields, fmt.Sprintf("%s=?", ITPublishTime))
		values = append(values, task.PublishTime)
	}

	if len(updateFields) > 0 {
		updateStr := fmt.Sprintf(
			"UPDATE %s SET %s WHERE %s=?",
			tblInspectTask,
			strings.Join(updateFields, ", "),
			fldID,
		)

		values = append(values, taskID)
		_, err = sql.Exec(updateStr, values...)
		if err != nil {
			err = fmt.Errorf("error while update inspect task in dao.Update, err: %s", err.Error())
			return
		}
	}
	return
}

func (dao *InspectTaskSqlDao) AssignInspectTasks(assigns []StaffTaskInfo, sql SqlLike) (err error) {
	if len(assigns) == 0 {
		return
	}
	valueStr := fmt.Sprintf("(?, ?, ?, ?)%s", strings.Repeat(", (?, ?, ?, ?)", len(assigns)-1))
	values := []interface{}{}
	for _, assign := range assigns {
		values = append(values, assign.TaskID, assign.StaffID, assign.CallID, assign.Type)
	}

	fields := []string{
		RITCSTaskID,
		RITCSStaffID,
		RITCSCallID,
		fldType,
	}
	fieldStr := strings.Join(fields, ", ")

	insertStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tblRelITCallStaff,
		fieldStr,
		valueStr,
	)

	_, err = sql.Exec(insertStr, values...)
	if err != nil {
		fmt.Errorf("error while assign inspect tasks in dao.AssignInspectTasks, err: %s", err.Error())
		return
	}
	return
}

func (dao *InspectTaskSqlDao) Outlines(sql SqlLike) (outlines []*Outline, err error) {
	queryStr := fmt.Sprintf(
		"SELECT %s, %s FROM %s",
		fldID,
		fldName,
		tblOutline,
	)

	rows, err := sql.Query(queryStr)
	if err != nil {
		logger.Error.Printf("error while query outlines in dao.Outlines, err: %s", err.Error())
		return
	}
	defer rows.Close()

	outlines = []*Outline{}
	for rows.Next() {
		outline := Outline{}
		rows.Scan(
			&outline.ID,
			&outline.Name,
		)
		outlines = append(outlines, &outline)
	}
	return
}

func (dao *InspectTaskSqlDao) UsersByType(userType string, sql SqlLike) (staffs []*Staff, err error) {
	queryStr := "SELECT uuid, display_name FROM users WHERE user_name Like '" + userType + "%'"

	rows, err := sql.Query(queryStr)
	if err != nil {
		logger.Error.Printf("error while query usrs in dao.UsersByType, err: %s", err.Error())
		return
	}
	defer rows.Close()

	staffs = []*Staff{}
	for rows.Next() {
		staff := Staff{}
		rows.Scan(
			&staff.UUID,
			&staff.Name,
		)
		staffs = append(staffs, &staff)
	}
	return
}

func (dao *InspectTaskSqlDao) FinishTask(staff string, callID int64, sql SqlLike) (err error) {
	updateStr := fmt.Sprintf(
		"UPDATE %s SET %s=1 WHERE %s=? and %s=?",
		tblRelITCallStaff,
		fldStatus,
		RITCSStaffID,
		RITCSCallID,
	)
	values := []interface{}{
		staff,
		callID,
	}

	logger.Info.Printf("values: %+v\n", values)

	_, err = sql.Exec(updateStr, values...)
	if err != nil {
		err = fmt.Errorf("error while set inspect task finished in dao.FinishTask, err: %s", err.Error())
		return
	}
	return
}

func (dao *InspectTaskSqlDao) ScoreForms(sql SqlLike) (forms []*ScoreForm, err error) {
	queryStr := fmt.Sprintf(
		"SELECT %s, %s FROM %s",
		fldID,
		fldName,
		tblScoreForm,
	)

	rows, err := sql.Query(queryStr)
	if err != nil {
		err = fmt.Errorf("error while get score forms in dao.ScoreForms, err: %s", err.Error())
		return
	}
	defer rows.Close()

	forms = []*ScoreForm{}
	for rows.Next() {
		form := ScoreForm{}
		rows.Scan(
			&form.ID,
			&form.Name,
		)
		forms = append(forms, &form)
	}
	return
}
