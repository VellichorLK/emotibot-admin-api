package model

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTasksSeed(t *testing.T) []Task {
	f, err := os.Open("./testdata/seed/task.csv")
	if err != nil {
		t.Fatal("can not open call's testdata, ", err)
	}
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal("can not read call's testdata ", err)
	}
	var tasks = make([]Task, 0, len(rows)-1)
	for _, row := range rows[1:] {
		var task Task
		Binding(&task, row)
		tasks = append(tasks, task)
	}
	return tasks
}
func TestITTaskDao_CallTask(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := TaskSQLDao{db: db}
	tasks := getTasksSeed(t)
	task, err := dao.CallTask(nil, Call{TaskID: 1})
	if err != nil {
		t.Fatal("expect call task query ok, but got ", err)
	}
	assert.Equal(t, tasks[0], task)

}

func TestITTaskDao_NewTask(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := TaskSQLDao{db: db}
	newTask := Task{
		Status:      2,
		IsDeal:      true,
		Series:      "123456",
		CreatedTime: 1546939580,
		UpdatedTime: 1546939580,
	}
	task, err := dao.NewTask(nil, newTask)
	if err != nil {
		t.Fatal("expect new task to be ok, but got ", err)
	}
	newTask.ID = task.ID
	assert.Equal(t, newTask, *task, "expect returned task should be exact as new task except for ID")
	fetchedTasks, err := dao.Task(nil, TaskQuery{
		ID: []int64{task.ID},
	})
	if err != nil {
		t.Fatal("expect query tasks to be ok, but got ", err)
	}
	assert.Equal(t, *task, fetchedTasks[0], "expect fetch task should be same as inserted task.")
}
