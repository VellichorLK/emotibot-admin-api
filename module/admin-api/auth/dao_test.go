package auth

import (
	"database/sql"
	"os"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var writer sqlmock.Sqlmock

func TestMain(m *testing.M) {
	var db *sql.DB
	db, writer, _ = sqlmock.New()
	util.SetDB(ModuleInfo.ModuleName, db)
	os.Exit(m.Run())
}
func TestGetUsername(t *testing.T) {
	var mockedRows = sqlmock.NewRows([]string{"uuid", "user_name"}).AddRow("0", "Dean").AddRow("1", "Daniel")
	writer.ExpectQuery("SELECT ").WithArgs("4b21158a395311e88a710242ac110003").WillReturnRows(mockedRows)
	usernames, err := GetUserNames([]string{"4b21158a395311e88a710242ac110003"})
	if err != nil {
		t.Fatal(err)
	}
	if len(usernames) != 2 {
		t.Fatal("user expect to be 2 but got ", len(usernames))
	}
	mockedRows = sqlmock.NewRows([]string{"enterprise"}).AddRow("APPLE")
	writer.ExpectQuery("SELECT enterprise FROM apps").WithArgs("csbot").WillReturnRows(mockedRows)
	mockedRows = sqlmock.NewRows([]string{"uuid", "user_name"}).AddRow("0", "Dean").AddRow("1", "Daniel")
	writer.ExpectQuery("SELECT uuid, user_name FROM users").WithArgs("APPLE").WillReturnRows(mockedRows)
	allnames, err := GetAllUserNames("csbot")
	if err != nil {
		t.Fatal(err)
	}
	if len(allnames) != 2 {
		t.Fatal("expect names to be 2 but got ", len(allnames))
	}
	if err = writer.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetUUID(t *testing.T) {
	var mockedRows = sqlmock.NewRows([]string{"uuid"}).AddRow("4b21158a395311e88a710242ac110003")
	writer.ExpectQuery("SELECT uuid FROM users").WithArgs("csbotadmin").WillReturnRows(mockedRows)
	uuid, err := GetUserID("csbotadmin")
	if err != nil {
		t.Fatal("GetUserID fail", err.Error())
	}
	if uuid != "4b21158a395311e88a710242ac110003" {
		t.Errorf("GetUserID except %+v, get %+v", "4b21158a395311e88a710242ac110003", uuid)
	}
	if err = writer.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
