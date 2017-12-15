package QA

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	httptest "github.com/kataras/iris/httptest"
)

type mockMCClient http.Client

func (m mockMCClient) McImportExcel(file multipart.FileHeader, UserID string, UserIP string, mode string) (util.MCResponse, error) {

	var mResp = util.MCResponse{
		SyncInfo: struct {
			StatID int    `json:"stateID"`
			UserID string `json:"userID"`
			Action string `json:"action"`
		}{
			123,
			"123",
			"123",
		},
	}
	return mResp, nil
}
func (m mockMCClient) McExportExcel(UserID string, UserIP string) (util.MCResponse, error) {
	return util.MCResponse{
		SyncInfo: struct {
			StatID int    `json:"stateID"`
			UserID string `json:"userID"`
			Action string `json:"action"`
		}{
			123,
			"123",
			"123",
		},
	}, nil
}

var app *iris.Application
var mHeader = map[string]string{
	"X-UserID":  "userX",
	"X-Real-IP": "0.0.0.0",
}

func TestMain(m *testing.M) {
	apiClient = mockMCClient{}
	app = iris.New()
	app.Post("/test/import", importExcel)
	app.Post("/test/export", exportExcel)
	app.Get("/test/{id:int}/download", download)
	app.Get("/test/{id:int}/progress", progress)
	app.Get("/test/views", viewOperations)
	//TODO Mock SQL DB
	if err := util.InitMainDB("127.0.0.1:3306", "root", "password", "emotibot"); err != nil {
		fmt.Println("Can not connect to local test server")
		os.Exit(1)
	}
	retCode := m.Run()
	os.Exit(retCode)
}
func TestImportExcel(t *testing.T) {
	e := httptest.New(t, app)
	var expectedResponse = "{\"state_id\":123}"
	var buf = strings.NewReader("test")
	body := e.POST("/test/import").WithMultipart().WithFormField("mode", "full_import").WithFile("file", "test.xlsx", buf).WithHeaders(mHeader).Expect().Status(200).Body().Equal(expectedResponse)
	//TODO Test different situaction and Audit Log
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
}

func TestExportExcel(t *testing.T) {
	e := httptest.New(t, app)

	var expectedResponse = "{\"state_id\":123}"
	body := e.POST("/test/export").WithHeaders(mHeader).Expect().Status(200).Body().Equal(expectedResponse)
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
}

func TestDownload(t *testing.T) {
	e := httptest.New(t, app)
	buf := bytes.NewBuffer([]byte{})
	util.LogInit(buf, buf, buf, buf)
	db := util.GetMainDB()
	rows, _ := db.Query("SELECT content, status FROM State_machine WHERE state_id = ?", 5)
	var expectedContent, status string
	if rows.Next() {
		rows.Scan(&expectedContent, &status)
		defer rows.Close()
	}
	//200
	body := e.GET("/test/{id}/download").WithPath("id", 5).WithHeaders(mHeader).Expect().Status(200).Body().Equal(expectedContent)
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
	//503
	e.GET("/test/{id}/download").WithPath("id", 4).WithHeaders(mHeader).Expect().Status(http.StatusServiceUnavailable)

}

func TestProgress(t *testing.T) {
	e := httptest.New(t, app)
	buf := bytes.NewBuffer([]byte{})
	util.LogInit(buf, buf, buf, buf)
	db := util.GetMainDB()
	rows, _ := db.Query("SELECT status FROM State_machine WHERE state_id = ?", 4)
	var expectedStatus string
	if rows.Next() {
		rows.Scan(&expectedStatus)
		defer rows.Close()
	}
	var mysqlLayout = "2006-01-02 15:04:05"

	selectedTime, err := time.Parse(mysqlLayout, "2017-12-13 08:27:26")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var successJSON, _ = json.Marshal(struct {
		ID          int          `json:"state_id"`
		Status      string       `json:"status"`
		CreatedTime JSONUnixTime `json:"created_time"`
		ExtraInfo   string       `json:"extra_info"`
	}{4, "running", JSONUnixTime(selectedTime), ""})

	body := e.GET("/test/{id}/progress").WithPath("id", 4).WithHeaders(mHeader).Expect().Status(200).Body().Equal(string(successJSON))
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
}

func TestViewOperations(t *testing.T) {
	e := httptest.New(t, app)
	values := url.Values{}
	// values.Set("userID", "dean")
	body := e.GET("/test/views").WithQueryString(values.Encode()).WithHeaders(mHeader).Expect().Status(200).Body()

	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}

}
