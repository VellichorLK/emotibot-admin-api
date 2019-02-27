package model

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedGroups create a slice of
func seedGroups(t *testing.T) []Group {
	f, err := os.Open("./testdata/seed/RuleGroup.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	//remove header
	records = records[1:]
	groups := make([]Group, 0, len(records))
	for i := 0; i < len(records); i++ {
		g := &Group{}
		Binding(g, records[i])
		groups = append(groups, *g)
	}

	return groups
}

func TestITGroupSQLDao_Group(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	seeds := seedGroups(t)
	type args struct {
		wantTx bool
		query  GroupQuery
	}
	queryEnt := "csbot"
	tests := []struct {
		name string
		args args
		want []Group
	}{
		{
			name: "get all",
			want: seeds[:2],
		},
		{
			name: "query enterprise",
			args: args{
				query: GroupQuery{EnterpriseID: &queryEnt, IgnoreSoftDelete: true},
			},
			want: []Group{seeds[2]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tx SqlLike
			if tt.args.wantTx {
				tx, _ = db.Begin()
			}
			groups, err := dao.Group(tx, tt.args.query)
			require.NoError(t, err)
			assert.Equal(t, tt.want, groups)
		})
	}

}
