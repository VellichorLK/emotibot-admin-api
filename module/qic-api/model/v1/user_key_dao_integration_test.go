package model

import (
	"encoding/csv"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func newUserKeys(t *testing.T) []UserKey {
	f, err := os.Open("./testdata/seed/UserKey.csv")
	if err != nil {
		t.Fatal("prepare test data failed, ", err)
	}
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatal("read prepared data failed, ", err)
	}
	var keys []UserKey
	for i := 1; i < len(records); i++ {
		var (
			j   = 0
			rec = records[i]
			key UserKey
			err error
		)
		key.ID, err = strconv.ParseInt(rec[j], 10, 64)
		if err != nil {
			t.Fatalf("parse col %d as id failed, %v", i, err)
		}
		j++
		key.Name = rec[j]
		j++
		key.Enterprise = rec[j]
		j++
		key.InputName = rec[j]
		j++
		typ, err := strconv.ParseInt(rec[j], 10, 8)
		if err != nil {
			t.Fatalf("parse col %d as type failed, %v", j, err)
		}
		key.Type = int8(typ)
		j++
		isdelete, err := strconv.ParseInt(rec[j], 10, 8)
		if err != nil {
			t.Fatalf("parse col %d as is_delete failed, %v", j, err)
		}
		if isdelete != 0 {
			key.IsDeleted = true
		}
		j++
		key.CreateTime, err = strconv.ParseInt(rec[j], 10, 64)
		if err != nil {
			t.Fatalf("parse col %d as create time failed %v", j, err)
		}
		j++
		key.UpdateTime, err = strconv.ParseInt(rec[j], 10, 64)
		if err != nil {
			t.Fatalf("parse col %d as update time failed, %v", j, err)
		}
		keys = append(keys, key)
	}
	return keys
}

func TestITUserKeySQLDaoUserKeys(t *testing.T) {
	dao := UserKeySQLDao{
		db: newIntegrationTestDB(t),
	}
	exampleKeys := newUserKeys(t)
	testTable := []struct {
		name         string
		arg          UserKeyQuery
		expectOutput []UserKey
		expectError  bool
	}{
		{
			name: "query id",
			arg: UserKeyQuery{
				ID: []int64{1},
			},
			expectOutput: exampleKeys[0:1],
		},
		{
			name: "query uuid",
			arg: UserKeyQuery{
				InputNames: []string{
					"location",
				},
			},
			expectOutput: exampleKeys[1:2],
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := dao.UserKeys(nil, tt.arg)
			if tt.expectError && err == nil {
				t.Fatal("expect error, but got nil error")
			}
			if err != nil && !tt.expectError {
				t.Fatal("not expect error, but got error ", err)
			}
			if !reflect.DeepEqual(keys, tt.expectOutput) {
				t.Logf("keys: %+v\nexpect:%+v\n", keys, tt.expectOutput)
				t.Error("non-equal output for result")
			}
		})
	}

}
