package model

import (
	"encoding/csv"
	"os"
	"reflect"
	"strconv"
	"testing"
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
	var tasks = make([]Task, len(rows)-1)
	for i, row := range rows[1:] {
		task := tasks[i]
		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		task.ID = id
		status, err := strconv.ParseInt(row[1], 10, 8)
		if err != nil {
			t.Fatal(err)
		}
		task.Status = int8(status)
		deal, err := strconv.ParseInt(row[3], 10, 8)
		if err != nil {
			t.Fatal(err)
		}
		if deal > 0 {
			task.IsDeal = true
		}
		task.Series = row[4]
		createdTime, err := strconv.ParseInt(row[5], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		task.CreatedTime = createdTime
		updatedTime, err := strconv.ParseInt(row[6], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		task.UpdatedTime = updatedTime
		tasks[i] = task
	}
	return tasks
}
func TestI11TaskDaoCallTask(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := TaskDao{db: db}
	tasks := getTasksSeed(t)
	task, err := dao.CallTask(nil, Call{TaskID: 1})
	if err != nil {
		t.Fatal("expect call task query ok, but got ", err)
	}
	if !reflect.DeepEqual(tasks[0], task) {
		t.Logf("expect task:\n%+v\noutput task:\n%+v\n", tasks[0], task)
		t.Error("expect task is not equal to output task")
	}
}

func TestITTaskDaoNewTask(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := TaskDao{db: db}
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
	if task.ID != 3 {
		t.Error("expect new task to insert id = 3, but got ", task.ID)
	}
	fetchedTasks, err := dao.Task(nil, TaskQuery{
		ID: []int64{task.ID},
	})
	if err != nil {
		t.Fatal("expect query tasks to be ok, but got ", err)
	}
	if !reflect.DeepEqual([]Task{*task}, fetchedTasks) {
		t.Logf("expect output:\n%+v\noutput:\n%+v", *task, fetchedTasks)
		t.Error("expect output task be exact with input task id reassigned")
	}

}
