package model

import (
	"testing"

	"logger"
	"math/rand"
	"time"
)

func TestKmeans(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	logger.ClsLogger = logger.SetLogPath("./")
	rawData := []Vector{
		{0, 2},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
		{6, 6},

		{100, 100},
		{101, 100},
		{102, 100},
		{103, 100},
		{104, 100},
		{105, 100},
		{106, 100},
	}

	labels := KmeansLabels(rawData, 2, 2)
	for i := 1; i < 7; i++ {
		if labels[0] != labels[i] {
			t.Error("fail_set_label0")
		}
	}

	for i := 8; i < 14; i++ {
		if labels[7] != labels[i] {
			t.Error("fail_set_label1")
		}
	}
}
