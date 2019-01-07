package model

import (
	_ "emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
)

type InspectTask struct {
	UUID              string
	Name              string
	Enterprise        string
	Description       string
	CallStart         int64
	CallEnd           int64
	Status            int
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
	ReviewCount       int
	ReviewTotal       int
	ReviewPercentage  int
	ReviewByPerson    int
	Staffs            []string
	ExcludeInspected  int8
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
			RITOOulineID,
		}

		valueStr := fmt.Sprintf("(?, ?)%s", strings.Repeat(", (?, ?)", outlineNum-1))
		values = make([]interface{}, 0)
		for _, outline := range task.Outlines {
			values = append(values, id, outline.ID)
		}

		insertStr = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s",
			tblRelITOuline,
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
