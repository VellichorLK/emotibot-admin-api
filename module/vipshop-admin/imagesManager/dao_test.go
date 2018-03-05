package imagesManager

import (
	"fmt"
	"os"
	"testing"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func TestMain(m *testing.M) {
	//
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	retCode := m.Run()
	os.Exit(retCode)

}
func TestGetLocationID(t *testing.T) {

	dao, err := util.InitDB("172.16.101.47:3306", "root", "password", "emotibot")
	if err != nil {
		t.Fatal(err)
	}

	util.SetDB(ModuleInfo.ModuleName, dao)

	id, err := getLocationID("http://vipshop/basemedia")
	if err != nil {
		t.Fatal(t)
	}
	fmt.Printf("id:%v\n", id)

}

func TestCreateBackupFolder(t *testing.T) {
	length := 10
	path := "."
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

	_, err = deleteFiles([]string{path + "/" + folerName})
	if err != nil {
		t.Fatal(err)
	}

}
