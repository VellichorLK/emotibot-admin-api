package imagesManager

import (
	"crypto/md5"
	"encoding/base64"
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
)

func TestReceiveImage(t *testing.T) {
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	var err error
	Volume, err = ioutil.TempDir("./testdata/", "image")
	if err != nil {
		t.Fatalf("Init failed, %v", err)
	}
	Volume += "/"
	LocalID = 1

	testFiles := []string{"golang.jpg", "test.txt"}
	fileArgs := make([]*uploadArg, len(testFiles))
	expectData := make([][]byte, len(testFiles))

	var mockedPic sqlmock.Sqlmock
	db, mockedPic, err = sqlmock.New()
	mockedPic.ExpectBegin()
	for i, f := range testFiles {
		data, err := ioutil.ReadFile("./testdata/" + f)
		if err != nil {
			t.Fatal(err)
		}
		arg := &uploadArg{FileName: f, Content: base64.StdEncoding.EncodeToString(data)}
		fileArgs[i] = arg
		expectData[i] = data
		stmt := mockedPic.ExpectPrepare("insert into images")
		stmt.ExpectExec().WillReturnResult(sqlmock.NewResult(int64(i), 1))
	}
	mockedPic.ExpectCommit()

	app := iris.New()
	app.Post("/test", receiveImage)
	e := httptest.New(t, app)
	e.POST("/test").WithJSON(fileArgs).Expect().Status(200)
	defer tearDownFiles(Volume)

	for i, f := range testFiles {
		d, err := ioutil.ReadFile(Volume + f)
		if err != nil {
			t.Fatal(err)
		}
		if md5.Sum(expectData[i]) != md5.Sum(d) {
			t.Error("Expect file %s should be the same after uploaded", f)
		}
	}
}
