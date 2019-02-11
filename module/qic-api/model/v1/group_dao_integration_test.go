package model

import (
	"encoding/csv"
	"os"
	"testing"
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
