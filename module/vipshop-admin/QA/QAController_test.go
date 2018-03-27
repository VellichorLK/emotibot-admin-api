package QA

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type mockMCClient http.Client

func (m mockMCClient) McImportExcel(file multipart.FileHeader, UserID string, UserIP string, mode string, appid string) (util.MCResponse, error) {

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
func (m mockMCClient) McExportExcel(UserID string, UserIP string, AnswerIDs []string, appid string, CategoryId int) (util.MCResponse, error) {
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

func (m mockMCClient) McManualBusiness(appid string) (int, error) {
	panic("Function not implemented")
}

var app *iris.Application
var mHeader = map[string]string{
	"X-UserID":  "userX",
	"X-Real-IP": "0.0.0.0",
	"Authorization": "vipshop",
}
var mainDBMock, auditDBMock sqlmock.Sqlmock

func TestMain(m *testing.M) {

	buf := bytes.NewBuffer([]byte{})
	util.LogInit(buf, buf, buf, buf)

	apiClient = mockMCClient{}

	app = iris.New()
	app.Post("/test/import", importExcel)
	app.Post("/test/export", exportExcel)
	app.Get("/test/{id:int}/download", download)
	app.Get("/test/{id:int}/progress", progress)
	app.Get("/test/views", viewOperations)
	var err error
	var db, auditDB *sql.DB
	db, mainDBMock, err = sqlmock.New()
	auditDB, auditDBMock, err = sqlmock.New()
	util.SetDB("main", db)
	util.SetDB("audit", auditDB)
	if err != nil {
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
	auditDBMock.ExpectExec("insert audit_record").WillReturnResult(sqlmock.NewResult(1, 1))
	//400 situaction
	e.POST("/test/import").WithMultipart().WithFormField("mode", "random").WithFile("file", "test.xlsx", buf).WithHeaders(mHeader).Expect().Status(http.StatusBadRequest)
	auditDBMock.ExpectExec("insert audit_record").WillReturnResult(sqlmock.NewResult(1, 1))
	e.POST("/test/import").WithMultipart().WithFormField("mode", "full").WithFile("file", "test.exe", buf).WithHeaders(mHeader).Expect().Status(http.StatusBadRequest)
	auditDBMock.ExpectExec("insert audit_record").WillReturnResult(sqlmock.NewResult(1, 1))
	body := e.POST("/test/import").WithMultipart().WithFormField("mode", "full").WithFile("file", "test.xlsx", buf).WithHeaders(mHeader).Expect().Status(200).Body().Equal(expectedResponse)
	//TODO Test different situaction and Audit Log
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
}

func TestExportExcel(t *testing.T) {
	e := httptest.New(t, app)

	var expectedResponse = "{\"state_id\":123}"
	var form = map[string]string {
		"category_id": "0",
		"timeset": "false",
		"search_question": "false",
		"search_answer": "false",
		"search_dm": "false",
		"search_rq": "false",
		"key_word": "",
		"not_show": "false",
		"page_limit": "10",
		"cur_page": "0",
		"search_all": "false",
		"dimension": "",
	}
	
	body := e.POST("/test/export").WithMultipart().WithForm(form).WithHeaders(mHeader).Expect().Status(200).Body().Equal(expectedResponse)
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
}

func TestDownload(t *testing.T) {
	e := httptest.New(t, app)
	id := "1"
	var expectedContent = "EMPTY"
	success := sqlmock.NewRows([]string{"content", "status"}).AddRow(expectedContent, "success")
	mainDBMock.ExpectQuery("SELECT content, status FROM state_machine").WithArgs(id).WillReturnRows(success)
	//200
	if body := e.GET("/test/{id}/download").WithPath("id", id).WithHeaders(mHeader).Expect().Status(200).Body().Equal(expectedContent); t.Failed() {
		fmt.Printf("Logging body: %s", body)
	}

	pending := sqlmock.NewRows([]string{"content", "status"}).AddRow([]byte{}, "running")
	mainDBMock.ExpectQuery("SELECT content, status FROM state_machine WHERE state_id").WithArgs(id).WillReturnRows(pending)
	//503
	if body := e.GET("/test/{id}/download").WithPath("id", id).WithHeaders(mHeader).Expect().Status(http.StatusServiceUnavailable); t.Failed() {
		fmt.Printf("Logging body: %s", body)
	}
}

func TestProgress(t *testing.T) {
	e := httptest.New(t, app)
	id := 1
	var expectedTime = JSONUnixTime{}
	rows := sqlmock.NewRows([]string{"status", "created_time", "extra_info"}).AddRow("running", expectedTime, "")
	mainDBMock.ExpectQuery("SELECT status, created_time, extra_info FROM state_machine  ").WithArgs(id).WillReturnRows(rows)
	var successJSON, _ = json.Marshal(struct {
		ID          int          `json:"state_id"`
		Status      string       `json:"status"`
		CreatedTime JSONUnixTime `json:"created_time"`
		ExtraInfo   string       `json:"extra_info"`
	}{1, "running", expectedTime, ""})

	body := e.GET("/test/{id}/progress").WithPath("id", id).WithHeaders(mHeader).Expect().Status(200).Body().Equal(string(successJSON))
	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}
}

func TestViewOperations(t *testing.T) {
	e := httptest.New(t, app)
	values := url.Values{}
	rows := sqlmock.NewRows([]string{"state_id", "user_id", "action", "status", "created_time", "updated_time", "extra_info"}).AddRow(1, "", "export", "success", nil, nil, "")
	mainDBMock.ExpectQuery("SELECT state_id, user_id, action, status, created_time, updated_time, extra_info FROM state_machine ").WillReturnRows(rows)
	body := e.GET("/test/views").WithQueryString(values.Encode()).WithHeaders(mHeader).Expect().Status(200).Body()

	if t.Failed() {
		fmt.Println("Logging Error message")
		fmt.Println(body)
	}

}
