package qi

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func TestTasksByCalls(t *testing.T) {
	/**
	*	setup mock
	**/
	tmp := callTask
	defer func() {
		callTask = tmp
	}()
	//dump dao only return a task with the call TaskID
	callTask = func(delegatee model.SqlLike, call model.Call) (model.Task, error) {
		t := model.Task{
			ID: call.TaskID,
		}
		// simulate some error cases
		if call.ID > 5 {
			return model.Task{}, errors.New("NOT FOUND")
		}
		return t, nil
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
					{ID: 6},
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
			// Since TasksByCalls has no guarantee of order, we have to compare it by ourself.
			for _, w := range tt.want {
				var found bool
				for _, g := range got {
					if reflect.DeepEqual(g, w) {
						found = true
					}
				}
				if !found {
					t.Errorf("want %v not in TasksByCalls", *w)
				}
			}
			t.Logf("TasksByCalls() = %s", prettyTasks(got))

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
