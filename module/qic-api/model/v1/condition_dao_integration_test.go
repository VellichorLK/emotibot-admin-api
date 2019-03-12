package model

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func readMockConditions(t *testing.T) []Condition {
	f, err := os.Open("./testdata/seed/RuleGroupCondition.csv")
	if err != nil {
		t.Fatal("read mock condition failed, ", err.Error())
	}
	defer f.Close()
	cf := csv.NewReader(f)
	records, err := cf.ReadAll()
	if err != nil {
		t.Fatalf("mock file csv read failed, %v", err)
	}
	var conditions []Condition
	for i := 1; i < len(records); i++ {
		var cond Condition
		rec := records[i]
		Binding(&cond, rec)
		conditions = append(conditions, cond)
	}
	return conditions
}
func TestConditionDao_Conditions(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	mc := readMockConditions(t)
	type args struct {
		query ConditionQuery
	}
	var tests = []struct {
		name    string
		args    args
		want    []Condition
		wantErr bool
	}{
		{
			name: "get all",
			args: args{},
			want: mc,
		},
		{
			name: "get single group's cond",
			args: args{
				query: ConditionQuery{
					GroupID: []int64{1},
				},
			},
			want: []Condition{mc[0]},
		},
		{
			name: "query type",
			args: args{
				query: ConditionQuery{
					Type: []int8{GroupCondTypOff},
				},
			},
			want: mc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := GroupConditionDao{db: &DefaultDBLike{DB: db}}
			got, err := dao.Conditions(nil, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, err = %v", tt.wantErr, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConditionDao_TestSuite(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupConditionDao{db: &DefaultDBLike{DB: db}}
	cond := Condition{
		GroupID: 666,
		Type:    GroupCondTypOn,
	}
	insertedCond, err := dao.NewCondition(nil, cond)
	require.NoError(t, err)
	fetchedCond, err := dao.Conditions(nil, ConditionQuery{ID: []int64{insertedCond.ID}})
	require.NoError(t, err)
	assert.Equal(t, fetchedCond[0], insertedCond)
}
