package model

import (
	"database/sql"
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

func newIntegrationTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1)/QISYS?parseTime=true&loc=Asia%2FTaipei")
	if err != nil {
		t.Fatal("expect db open success but got error: ", err)
	}
	return db
}
