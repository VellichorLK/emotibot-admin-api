package model

import (
	"flag"
	"os"
	"testing"
)

var isIntegration bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegration, "integration", false, "flag for running integration test")
	flag.Parse()
	os.Exit(m.Run())
}

func skipIntergartion(t *testing.T) {
	if !isIntegration {
		t.Skip("skip intergration test, please specify -intergation flag.")
	}
	return
}
