package model

import (
	"encoding/csv"
	"os"
	"reflect"
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
		key := &UserKey{}
		Binding(key, records[i])
		keys = append(keys, *key)
	}
	return keys
}

func TestITUserKeySQLDaoUserKeys(t *testing.T) {
	skipIntergartion(t)
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
		{
			name: "ignore soft deleted",
			arg: UserKeyQuery{
				IgnoreSoftDelete: true,
			},
			expectOutput: exampleKeys,
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
