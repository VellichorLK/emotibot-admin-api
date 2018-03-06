package imagesManager

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"testing"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
)

func TestReceiveImage(t *testing.T) {
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	url := "172.16.101.47:3306"
	user := "root"
	pass := "password"
	dbName := "emotibot"

	dao, err := util.InitDB(url, user, pass, dbName)
	if err != nil {
		util.LogError.Printf("Cannot init self learning db, [%s:%s@%s:%s]: %s\n", user, pass, url, dbName, err.Error())
		return
	}
	//util.SetDB(ModuleInfo.ModuleName, dao)
	db = dao
	Volume = "./"
	LocalID = 1

	app := iris.New()
	app.Post("/test", receiveImage)
	e := httptest.New(t, app)

	data1, err := ioutil.ReadFile("./testdata/golang.jpg")
	if err != nil {
		t.Fatal()
	}
	file1 := &uploadArg{FileName: "myfile1.jpg", Content: base64.StdEncoding.EncodeToString(data1)}

	data2, err := ioutil.ReadFile("./controller_test.go")
	if err != nil {
		t.Fatal()
	}
	file2 := &uploadArg{FileName: "controller_test22.go", Content: base64.StdEncoding.EncodeToString(data2)}

	files := make([]*uploadArg, 2, 2)
	files[0] = file1
	files[1] = file2

	e.POST("/test").WithJSON(files).Expect().Status(200)

}
