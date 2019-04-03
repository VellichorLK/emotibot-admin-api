package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_whereBuilder_Between(t *testing.T) {
	type fields struct {
		ConcatLogic boolLogic
		alias       string
		data        []interface{}
		conditions  []string
	}
	type args struct {
		fieldName string
		rangeCond RangeCondition
	}
	newInt64 := func(i int64) *int64 {
		return &i
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSQL  string
		wantData []interface{}
	}{
		{
			name: "between",
			fields: fields{
				alias: "a.",
			},
			args: args{
				fieldName: "timestamp",
				rangeCond: RangeCondition{
					lb: newInt64(1),
					ub: newInt64(10),
				},
			},
			wantSQL:  "a.`timestamp` BETWEEN ? AND ?",
			wantData: []interface{}{int64(1), int64(10)},
		},
		{
			name: "Gte",
			fields: fields{
				alias: "a.",
			},
			args: args{
				fieldName: "timestamp",
				rangeCond: RangeCondition{
					lb: newInt64(1),
				},
			},
			wantSQL:  "a.`timestamp` >= ?",
			wantData: []interface{}{int64(1)},
		},
		{
			name: "Lte",
			fields: fields{
				alias: "a.",
			},
			args: args{
				fieldName: "timestamp",
				rangeCond: RangeCondition{
					ub: newInt64(10),
				},
			},
			wantSQL:  "a.`timestamp` <= ?",
			wantData: []interface{}{int64(10)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &whereBuilder{
				ConcatLogic: tt.fields.ConcatLogic,
				alias:       tt.fields.alias,
				data:        tt.fields.data,
				conditions:  tt.fields.conditions,
			}
			w.Between(tt.args.fieldName, tt.args.rangeCond)
			if tt.wantSQL == "" && len(w.conditions) == 0 {
				return
			}
			assert.Equal(t, tt.wantSQL, w.conditions[len(w.conditions)-1])
			assert.Equal(t, tt.wantData, w.data)
		})
	}
}

func TestEscapeLike(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "escape percentage",
			args: args{
				query: "%HI%",
			},
			want: `\%HI\%`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeLike(tt.args.query); got != tt.want {
				t.Errorf("EscapeLike() = %v, want %v", got, tt.want)
			}
		})
	}
}
