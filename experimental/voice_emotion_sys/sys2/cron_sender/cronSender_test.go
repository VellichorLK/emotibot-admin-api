package main

import (
	"fmt"
	"handlers"
	"testing"
	"time"
)

func TestGetQueryTime(t *testing.T) {
	base := time.Now()
	qt, err := getQueryTime(&base, handlers.DAY)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Printf("from:%s\nto:%s\nlastFrom:%s\nlastTo:%s\n", time.Unix(qt.From, 0), time.Unix(qt.To, 0), time.Unix(qt.LastFrom, 0), time.Unix(qt.LastTo, 0))
	}
}
