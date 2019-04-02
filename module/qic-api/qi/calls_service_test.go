package qi

import (
	"testing"

	"github.com/stretchr/testify/require"

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

func Test_matchDefaultConditions(t *testing.T) {
	type args struct {
		groups []model.Group
		call   model.Call
	}
	tests := []struct {
		name            string
		args            args
		filteredIndexes []int // must be descending order index
	}{
		{
			name: "test type 1 group condition",
			args: args{
				groups: []model.Group{
					{
						ID: 1,
						Condition: &model.Condition{
							Type: 1,
						},
					},
				},
				call: model.Call{},
			},
			filteredIndexes: []int{},
		},
		{
			name: "filter UploadTime range",
			args: args{
				groups: []model.Group{
					{
						ID: 1,
						Condition: &model.Condition{
							UploadTimeStart: 1553159400,
						},
					},
					{
						ID: 2,
						Condition: &model.Condition{
							UploadTimeEnd: 1553159600,
						},
					},
					{
						ID: 3,
						Condition: &model.Condition{
							UploadTimeStart: 1553159600,
						},
					},
					{
						ID: 4,
						Condition: &model.Condition{
							UploadTimeEnd: 1553159000,
						},
					},
					{
						ID: 5,
						Condition: &model.Condition{
							UploadTimeStart: 1553159300,
							UploadTimeEnd:   1553159600,
						},
					},
				},
				call: model.Call{
					UploadUnixTime: 1553159500,
				},
			},
			filteredIndexes: []int{3, 2},
		},
		{
			name: "filter call time range",
			args: args{
				groups: []model.Group{
					{
						ID: 1,
						Condition: &model.Condition{
							CallStart: 1553159400,
						},
					},
					{
						ID: 2,
						Condition: &model.Condition{
							CallEnd: 1553159600,
						},
					},
					{
						ID: 3,
						Condition: &model.Condition{
							CallStart: 1553159600,
						},
					},
					{
						ID: 4,
						Condition: &model.Condition{
							CallEnd: 1553159000,
						},
					},
					{
						ID: 5,
						Condition: &model.Condition{
							CallStart: 1553159300,
							CallEnd:   1553159600,
						},
					},
				},
				call: model.Call{
					CallUnixTime: 1553159500,
				},
			},
			filteredIndexes: []int{3, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := tt.args.groups
			got := matchDefaultConditions(tt.args.groups, tt.args.call)
			for _, idx := range tt.filteredIndexes {
				groups = append(groups[:idx], groups[idx+1:]...)
			}
			assert.Equal(t, groups, got)
		})
	}
}

func TestCallRespoWithTotal(t *testing.T) {
	defer BackupPointers(&callCount, &calls, &valuesKey)()
	type args struct {
		query model.CallQuery
	}
	type want struct {
		responses []CallResp
		total     int64
	}
	type mock struct {
		callCount func(delegatee model.SqlLike, query model.CallQuery) (int64, error)
		calls     func(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error)
		valuesKey func(delegatee model.SqlLike, query model.UserValueQuery) ([]model.UserValue, error)
	}
	tests := []struct {
		name string
		args args
		want want
		mock
		wantErr bool
	}{
		{
			name: "not found",
			want: want{
				responses: []CallResp{},
			},
			mock: mock{
				callCount: func(delegatee model.SqlLike, query model.CallQuery) (int64, error) {
					return 0, nil
				},
				calls: func(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error) {
					return []model.Call{}, nil
				},
				valuesKey: func(delegatee model.SqlLike, query model.UserValueQuery) ([]model.UserValue, error) {
					return []model.UserValue{}, nil
				},
			},
		},
		{
			name: "custom columns",
			mock: mock{
				callCount: func(delegatee model.SqlLike, query model.CallQuery) (int64, error) {
					return 1, nil
				},
				calls: func(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error) {
					var empty string
					return []model.Call{
						model.Call{
							UUID:          "de67910159d84d25ae51b2455332a419",
							ID:            1,
							FileName:      &empty,
							Description:   &empty,
							LeftChanRole:  model.CallChanStaff,
							RightChanRole: model.CallChanCustomer,
						},
					}, nil
				},
				valuesKey: func(delegatee model.SqlLike, query model.UserValueQuery) ([]model.UserValue, error) {
					values := []model.UserValue{
						model.UserValue{
							Type:      model.UserValueTypCall,
							LinkID:    1,
							UserKeyID: 1,
							Value:     "hello",
							UserKey: &model.UserKey{
								ID:        1,
								Type:      model.UserKeyTypArray,
								InputName: "mycol",
							},
						},
						model.UserValue{
							Type:      model.UserValueTypCall,
							LinkID:    1,
							UserKeyID: 1,
							Value:     "world",
							UserKey: &model.UserKey{
								ID:        1,
								Type:      model.UserKeyTypArray,
								InputName: "mycol",
							},
						},
						model.UserValue{
							Type:      model.UserValueTypCall,
							LinkID:    1,
							UserKeyID: 1,
							Value:     "1",
							UserKey: &model.UserKey{
								ID:        1,
								Type:      model.UserKeyTypNumber,
								InputName: "mycolNum",
							},
						},
						model.UserValue{
							Type:      model.UserValueTypCall,
							LinkID:    1,
							UserKeyID: 1,
							Value:     "demo",
							UserKey: &model.UserKey{
								ID:        1,
								Type:      model.UserKeyTypString,
								InputName: "mycolStr",
							},
						},
					}
					return values, nil
				},
			},
			want: want{
				responses: []CallResp{
					CallResp{
						CallID:       1,
						CallUUID:     "de67910159d84d25ae51b2455332a419",
						LeftChannel:  "staff",
						RightChannel: "customer",
						CustomColumns: map[string]interface{}{
							"mycol":    []string{"hello", "world"},
							"mycolNum": 1,
							"mycolStr": "demo",
						},
					},
				},
				total: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount = tt.mock.callCount
			calls = tt.mock.calls
			valuesKey = tt.mock.valuesKey
			gotResp, total, err := CallRespsWithTotal(tt.args.query)
			require.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, want{
				responses: gotResp,
				total:     total,
			})
		})
	}
}
