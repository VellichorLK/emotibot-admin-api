package model

import (
	"encoding/csv"
	"os"
	"testing"
)

func getCallsSeed(t *testing.T) []Call {
	f, err := os.Open("./testdata/seed/call.csv")
	if err != nil {
		t.Fatal("can not open call's testdata, ", err)
	}
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal("can not read call's testdata ", err)
	}
	var calls = make([]Call, 0)
	for i := len(rows[1:]); i >= 1; i-- {
		var c Call
		Binding(&c, rows[i])
		calls = append(calls, c)
	}
	return calls
}
