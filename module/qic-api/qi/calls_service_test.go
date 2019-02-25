package qi

import (
	"testing"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"github.com/stretchr/testify/assert"
)

func Test_insertAndOrderStrings(t *testing.T) {
	type args struct {
		orderedValues []string
		v             string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal",
			args: args{
				orderedValues: []string{"a", "b", "d"},
				v:             "c",
			},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "empty insert",
			args: args{
				orderedValues: []string{},
				v:             "a",
			},
			want: []string{"a"},
		},
		{
			name: "insert at tail",
			args: args{
				orderedValues: []string{"a", "b", "c"},
				v:             "d",
			},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "insert at head",
			args: args{
				orderedValues: []string{"b", "c", "d"},
				v:             "a",
			},
			want: []string{"a", "b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, insertAndOrderStrings(tt.args.orderedValues, tt.args.v))
		})
	}
}

func TestMatchGroup(t *testing.T) {
	type args struct {
		groups          []model.Group
		groupConditions map[int64][]model.UserKey
		userInputs      map[string][]string
	}
	tests := []struct {
		name string
		args args
		want []model.Group
	}{
		{
			name: "empty condition", // any empty condition group should be consider success
			args: args{
				groups: []model.Group{
					{ID: 1}, {ID: 2},
				},
				groupConditions: map[int64][]model.UserKey{
					1: []model.UserKey{
						model.UserKey{ID: 1, InputName: "loc", UserValues: []model.UserValue{{Value: "taipei"}}},
					},
				},
				userInputs: map[string][]string{
					"loc": []string{"taipei", "taichung"},
				},
			},
			want: []model.Group{{ID: 1}, {ID: 2}},
		},
		{
			name: "condition filtering",
			args: args{
				groups: []model.Group{
					{ID: 1}, {ID: 2},
				},
				groupConditions: map[int64][]model.UserKey{
					1: []model.UserKey{
						{ID: 1, InputName: "area", UserValues: []model.UserValue{{Value: "CHINA"}}},
						{ID: 2, InputName: "LANG", UserValues: []model.UserValue{{Value: "EN"}}},
					},
					2: []model.UserKey{
						{ID: 1, InputName: "area", UserValues: []model.UserValue{{Value: "CHINA"}, {Value: "USA"}}},
					},
				},
				userInputs: map[string][]string{
					"area": []string{"USA"},
				},
			},
			want: []model.Group{
				{ID: 2},
			},
		},
		{
			name: "empty values",
			args: args{
				groups: []model.Group{
					{ID: 3},
				},
				groupConditions: map[int64][]model.UserKey{
					3: []model.UserKey{
						{InputName: "area"},
					},
				},
				userInputs: map[string][]string{
					"area": []string{"USA"},
				},
			},
			want: []model.Group{},
		},
		{
			name: "wrong name",
			args: args{
				groups: []model.Group{
					{ID: 1},
				},
				groupConditions: map[int64][]model.UserKey{
					1: []model.UserKey{
						{InputName: "area", UserValues: []model.UserValue{{Value: "USA"}}},
					},
				},
				userInputs: map[string][]string{
					"area-wrongname": []string{"USA"},
				},
			},
			want: []model.Group{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchGroup(tt.args.groups, tt.args.groupConditions, tt.args.userInputs)
			assert.Equal(t, tt.want, got)
		})
	}
}
