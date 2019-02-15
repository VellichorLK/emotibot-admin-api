package model

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"
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
