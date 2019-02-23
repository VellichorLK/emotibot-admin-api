package model

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readMockUserValues(t *testing.T) []UserValue {
	f, err := os.Open("./testdata/seed/UserValue.csv")
	if err != nil {
		t.Fatal("prepare test data failed, ", err)
	}
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatal("read prepared data failed, ", err)
	}
	var values []UserValue
	for i := 1; i < len(records); i++ {
		v := UserValue{}
		Binding(&v, records[i])
		values = append(values, v)
	}
	return values
}
func TestITUserValueDaoUserValues(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := UserValueDao{conn: db}
	examples := readMockUserValues(t)
	type args struct {
		delegatee SqlLike
		query     UserValueQuery
	}
	var testTable = []struct {
		name    string
		args    args
		want    []UserValue
		wantErr bool
	}{
		{
			name: "ID",
			args: args{
				query: UserValueQuery{
					ID: []int64{1},
				},
			},
			want: examples[:1],
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.UserValues(tt.args.delegatee, tt.args.query)
			if (err == nil) == tt.wantErr {
				t.Fatalf("UserValues err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserValues got %v, want = %v", got, tt.want)
			}
		})
	}

}

func TestITUserValueDaoTestSuite(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := UserValueDao{conn: db}
	example := UserValue{
		UserKeyID: 1,
	}
	got, err := dao.NewUserValue(nil, example)
	if err != nil {
		t.Fatalf("NewUserValue error = %v", err)
	}
	query := UserValueQuery{ID: []int64{got.ID}}
	written, err := dao.UserValues(nil, query)
	if err != nil {
		t.Fatalf("UserValues err = %v", err)
	}
	if !reflect.DeepEqual(got, written[0]) {
		t.Errorf("written = %+v, got = %+v", written[0], got)
	}
	total, err := dao.DeleteUserValues(nil, query)
	if err != nil {
		t.Fatalf("DeleteUserValues err = %v", err)
	}
	if total != 1 {
		t.Errorf("expect delete 1 row, but total = %d", total)
	}

}

func TestITUserValueDao_ValuesKey(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := UserValueDao{conn: db}
	type args struct {
		delegatee SqlLike
		query     UserValueQuery
	}
	uservalues := readMockUserValues(t)
	userkeys := newUserKeys(t)
	for i, v := range uservalues {
		for _, k := range userkeys {
			if k.ID == v.UserKeyID {
				v.UserKey = &k
				uservalues[i] = v
				break
			}
		}
	}
	var tests = []struct {
		name    string
		args    args
		want    []UserValue
		wantErr bool
	}{
		{
			name: "query id",
			args: args{
				query: UserValueQuery{
					ID: []int64{1, 2, 3},
				},
			},
			want: uservalues[:3],
		},
		{
			name: "query type",
			args: args{
				query: UserValueQuery{
					ParentID: []int64{1},
					Type:     []int8{UserValueTypCall},
				},
			},
			want: uservalues[1:4],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.ValuesKey(tt.args.delegatee, tt.args.query)
			if tt.wantErr == (err == nil) {
				t.Fatalf("wantErr = %v, err = %v", tt.wantErr, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
