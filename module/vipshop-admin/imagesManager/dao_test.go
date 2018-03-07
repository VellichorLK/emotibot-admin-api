package imagesManager

import (
	"os"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func TestMain(m *testing.M) {
	os.Setenv("SYNC_PERIOD_BY_SECONDS", "100")
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	retCode := m.Run()
	os.Exit(retCode)

}
func TestGetLocationID(t *testing.T) {
	var mockedImg sqlmock.Sqlmock
	db, mockedImg, _ = sqlmock.New()
	input := "http://vipshop/basemedia"
	stmt := mockedImg.ExpectPrepare("insert into")
	stmt.ExpectExec().WithArgs(input).WillReturnResult(sqlmock.NewResult(1, 1))
	id, err := getLocationID(input)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Errorf("expect id = 1, but got %v\n", id)
	}

}

func TestCreateBackupFolder(t *testing.T) {
	length := 10
	path := "./testdata"
	folder, err := createBackupFolder(length, path)
	if err != nil {
		t.Fatal(err)
	}

	if length != len(folder) {
		t.Fatalf("folder lenght %d is not %d,", len(folder), length)
	}

	var file os.FileInfo

	if file, err = os.Stat(path + "/" + folder); err != nil {
		t.Fatal(err)
	}

	if !file.IsDir() {
		t.Fatal("create folder is not folder")
	}

	if file.Name() != folder {
		t.Fatalf("created foler %s is not assigned name %s\n", file.Name(), folder)
	}
	tearDownFiles(path + "/" + folder)

}

func TestBackupFilesFlow(t *testing.T) {
	length := 10
	path := "."
	folerName, err := createBackupFolder(length, path)
	if err != nil {
		t.Fatal(err)
	}

	_, err = copyFiles(path, path+"/"+folerName, []string{"controller.go", "dao.go", "dao_test.go"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = deleteFiles("", []string{path + "/" + folerName})
	if err != nil {
		t.Fatal(err)
	}

}
