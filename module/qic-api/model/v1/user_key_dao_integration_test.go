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
		expectCount  int64
		expectError  bool
	}{
		{
			name: "query id",
			arg: UserKeyQuery{
				ID: []int64{1},
			},
			expectOutput: exampleKeys[0:1],
			expectCount:  1,
		},
		{
			name: "query uuid",
			arg: UserKeyQuery{
				InputNames: []string{
					"location",
				},
			},
			expectOutput: exampleKeys[1:2],
			expectCount:  1,
		},
		{
			name: "query fuzzy name",
			arg: UserKeyQuery{
				FuzzyName: "地",
			},
			expectOutput: exampleKeys[1:2],
			expectCount:  1,
		},
		{
			name: "query with wildcard char",
			arg: UserKeyQuery{
				FuzzyName: "%",
			},
			expectOutput: []UserKey{},
			expectCount:  0,
		},
		{
			name: "ignore soft deleted",
			arg: UserKeyQuery{
				IgnoreSoftDelete: true,
			},
			expectOutput: []UserKey{exampleKeys[2], exampleKeys[0], exampleKeys[1]},
			expectCount:  int64(len(exampleKeys)),
		},
		{
			name: "query with pagination",
			arg: UserKeyQuery{
				IgnoreSoftDelete: true,
				Paging: &Pagination{
					Limit: 1,
					Page:  1,
				},
			},
			expectOutput: exampleKeys[2:],
			expectCount:  int64(len(exampleKeys)),
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
				t.Logf("keys: %#v\nexpect:%#v\n", keys, tt.expectOutput)
				t.Error("non-equal output for result")
			}
			total, err := dao.CountUserKeys(nil, tt.arg)
			if err != nil {
				t.Fatal("expect count ok, but got ", err)
			}
			if tt.expectCount != total {
				t.Errorf("expect count to be %d, but got %d", tt.expectCount, total)
			}
		})
	}

}

func TestITUserKeySQLDaoNewUserKey(t *testing.T) {
	dao := UserKeySQLDao{
		db: newIntegrationTestDB(t),
	}
	key := UserKey{
		Name:       "地區",
		Enterprise: "832ec03d470b49dab3a0f017bf27ff45",
		InputName:  "c0032b4d3aa142a09e5ea10893707e7c",
		Type:       UserKeyTypString,
		IsDeleted:  false,
		CreateTime: 1549857200,
		UpdateTime: 1549857300,
	}
	createdKey, err := dao.NewUserKey(nil, key)
	if err != nil {
		t.Fatal("expect ok, but got ", err)
	}
	if createdKey.ID == 0 {
		t.Fatal("expect key id to be assigned, but got zero")
	}
	key.ID = createdKey.ID
	if !reflect.DeepEqual(key, createdKey) {
		t.Logf("request: %+v\ncreated: %+v\n", key, createdKey)
		t.Error("expect created key to be the same, but not equal")
	}
	query := UserKeyQuery{
		ID: []int64{createdKey.ID},
	}
	keys, err := dao.UserKeys(nil, query)
	if err != nil {
		t.Fatal("expect query ok, but got ", err)
	}
	if len(keys) != 1 {
		t.Fatal("expect one element found, but got ", len(keys))
	}
	if !reflect.DeepEqual(keys[0], createdKey) {
		t.Logf("keys[1]: %+v\nCreated: %+v\n", keys[0], createdKey)
		t.Error("expect created key to be the same as keys 1")
	}
}
