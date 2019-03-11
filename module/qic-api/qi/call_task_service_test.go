package qi

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func TestTasksByCalls(t *testing.T) {
	/**
	*	setup mock
	**/
	tmp := tasks
	defer func() {
		tasks = tmp
	}()
	tasks = func(delegatee model.SqlLike, query model.TaskQuery) ([]model.Task, error) {
		var taskGrp = make([]model.Task, 0)
		sort.Slice(query.ID, func(i, j int) bool {
			if query.ID[i] > query.ID[j] {
				return false
			} else {
				return true
			}
		})
		for _, id := range query.ID {
			taskGrp = append(taskGrp, model.Task{
				ID: id,
			})
			if id > 5 {
				return nil, errors.New("NOT FOUND")
			}
		}
		return taskGrp, nil
	}
	/**
	*	test table
	**/
	type args struct {
		calls []model.Call
	}
	tests := []struct {
		name    string
		args    args
		want    []*model.Task
		wantErr bool
	}{
		{
			name: "calls with different staff",
			args: args{
				calls: []model.Call{
					{ID: 1, TaskID: 1, StaffID: "dean"}, {ID: 2, TaskID: 1, StaffID: "taylor"},
				},
			},
			want: []*model.Task{
				{
					ID: 1,
					CallsOfStaffs: map[string][]model.Call{
						"dean":   []model.Call{model.Call{ID: 1, TaskID: 1, StaffID: "dean"}},
						"taylor": []model.Call{model.Call{ID: 2, TaskID: 1, StaffID: "taylor"}},
					},
				},
			},
		},
		{
			name: "calls with same staff",
			args: args{
				calls: []model.Call{
					{ID: 1, TaskID: 1, StaffID: "joanne"}, {ID: 2, TaskID: 1, StaffID: "joanne"},
				},
			},
			want: []*model.Task{
				{
					ID: 1,
					CallsOfStaffs: map[string][]model.Call{
						"joanne": []model.Call{{ID: 1, TaskID: 1, StaffID: "joanne"}, {ID: 2, TaskID: 1, StaffID: "joanne"}},
					},
				},
			},
		},
		{
			name: "calls with different tasks",
			args: args{
				calls: []model.Call{
					{ID: 1, TaskID: 1, StaffID: "chester"}, {ID: 2, TaskID: 2, StaffID: "elaine"},
				},
			},
			want: []*model.Task{
				{
					ID: 1,
					CallsOfStaffs: map[string][]model.Call{
						"chester": []model.Call{{ID: 1, TaskID: 1, StaffID: "chester"}},
					},
				},
				{
					ID: 2,
					CallsOfStaffs: map[string][]model.Call{
						"elaine": []model.Call{{ID: 2, TaskID: 2, StaffID: "elaine"}},
					},
				},
			},
		},
		{
			name: "non exist call",
			args: args{
				calls: []model.Call{
					{ID: 1, TaskID: 6},
				},
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	/**
	*	test code
	**/
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TasksByCalls(tt.args.calls)
			if (err != nil) != tt.wantErr {
				t.Errorf("TasksByCalls() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func prettyTasks(tasks []*model.Task) string {
	msg := "{"
	for _, t := range tasks {
		var c string
		for id, calls := range t.CallsOfStaffs {
			c = fmt.Sprintf("%s %s: [%s]", c, id, prettyCalls(calls))
		}
		c = fmt.Sprintf("[%s]", c)
		msg = fmt.Sprintf("%s Task{ID:%d, Staffs: %v},", msg, t.ID, c)

	}
	msg += "}"
	return msg
}
func prettyCalls(calls []model.Call) string {
	msg := "{"
	for _, c := range calls {
		msg = fmt.Sprintf("%s {ID: %d, StaffID: %s}", msg, c.ID, c.StaffID)
	}
	msg += "}"
	return msg
}
