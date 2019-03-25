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

func TestITTaskDao_Task(t *testing.T) {
	skipIntergartion(t)
	tasks := getTasksSeed(t)
	type args struct {
		delegatee SqlLike
		query     TaskQuery
	}
	var tests = []struct {
		name    string
		args    args
		want    []Task
		wantErr bool
	}{
		{
			name: "query ID",
			args: args{
				query: TaskQuery{
					ID: []int64{2},
				},
			},
			want: []Task{tasks[1]},
		},
		{
			name: "query sn",
			args: args{
				query: TaskQuery{
					SN: []string{"47960"},
				},
			},
			want: []Task{tasks[0]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newIntegrationTestDB(t)
			dao := TaskSQLDao{db: db}
			got, err := dao.Task(tt.args.delegatee, tt.args.query)
			if tt.wantErr == (err == nil) {
				t.Fatalf("wantErr = %v, err = %v", tt.wantErr, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
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
	checkDBStat(t)
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
	checkDBStat(t)
}