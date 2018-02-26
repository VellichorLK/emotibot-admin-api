package FAQ

import (
	"bytes"
	"database/sql"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var app *iris.Application
var mHeader = map[string]string{
	"Authorization": "vipshop",
	"X-UserID":      "userX",
	"X-Real-IP":     "0.0.0.0",
}

var mockedMainDB sqlmock.Sqlmock
var mockedAuditDB sqlmock.Sqlmock

type mockedMCClient struct{}

func (mockedMCClient) McImportExcel(fileHeader multipart.FileHeader, UserID string, UserIP string, mode string) (util.MCResponse, error) {
	panic("function is not impelmented")
}
func (mockedMCClient) McExportExcel(UserID string, UserIP string, answerIDs []string) (util.MCResponse, error) {
	panic("function is not impelmented")
}
func (mockedMCClient) McManualBusiness(appid string) (int, error) {
	return 0, nil
}

var errBuff = bytes.NewBuffer(make([]byte, 100))

func TestMain(m *testing.M) {

	util.LogInit(os.Stdout, os.Stdout, os.Stdout, errBuff)

	app = iris.New()
	app.Post("/question/{qid:string}/similar-questions", handleUpdateSimilarQuestions)
	var err error
	var db, auditDB *sql.DB
	db, mockedMainDB, err = sqlmock.New()
	if err != nil {
		os.Exit(1)
	}

	auditDB, mockedAuditDB, err = sqlmock.New()
	if err != nil {
		os.Exit(1)
	}
	util.SetDB("main", db)
	util.SetDB("audit", auditDB)

	retCode := m.Run()
	os.Exit(retCode)
}

//TODO: Multicustomer mock up neeeded!
func TestUpdateHandlerAuditLog(t *testing.T) {
	testCases := []struct {
		Scenario               string
		Body                   []SimilarQuestion
		ExpectedStatusCode     int
		ExpectedAuditOperation string
		ExpectedAuditLog       string
	}{
		{
			"更新",
			[]SimilarQuestion{
				SimilarQuestion{"相似问题2", "1"},
				SimilarQuestion{"相似问题3", "2"},
				SimilarQuestion{"相似问题4", "3"},
			},
			200,
			util.AuditOperationEdit,
			//預期audit log結果為：相似問題1 => 2,3  4號因為兩邊都有所以移除了
			"[相似问题]:[/LEVEL1/LEVEL2/LEVEL3][标准问题1]:相似问题1=>相似问题2;相似问题3",
		},
		{
			"新增",
			[]SimilarQuestion{
				SimilarQuestion{"相似问题1", "1"},
				SimilarQuestion{"相似问题3", "2"},
				SimilarQuestion{"相似问题4", "3"},
			},
			200,
			util.AuditOperationAdd,
			"[相似问题]:[/LEVEL1/LEVEL2/LEVEL3][标准问题1]:相似问题3",
		},
		{
			"刪除",
			[]SimilarQuestion{
				SimilarQuestion{"相似问题1", "1"},
			},
			200,
			util.AuditOperationDelete,
			"[相似问题]:[/LEVEL1/LEVEL2/LEVEL3][标准问题1]:相似问题4",
		},
	}
	e := httptest.New(t, app)
	qid := 123
	util.DefaultMCClient = mockedMCClient{}

	for _, tt := range testCases {
		t.Run(tt.Scenario, func(t *testing.T) {
			errBuff.Reset()
			//設定場景: /LEVEL1/LEVEL2/LEVEL3 之 標準問題1，有相似問題1,4。 更新相似問 2,3,4。
			qRows := sqlmock.NewRows([]string{"Content", "Category_Id"}).AddRow("标准问题1", "3")
			cRows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(1, "LEVEL1", 0).AddRow(2, "LEVEL2", 1).AddRow(3, "LEVEL3", 2)
			sRows := sqlmock.NewRows([]string{"Content"}).AddRow("相似问题1").AddRow("相似问题4")
			mockedMainDB.ExpectQuery("SELECT Content, Category_Id from vipshop_question").WillReturnRows(qRows)
			mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(cRows)
			mockedMainDB.ExpectQuery("SELECT Content FROM vipshop_squestion").WillReturnRows(sRows)
			mockedMainDB.ExpectBegin()
			mockedMainDB.ExpectExec("DELETE FROM vipshop_squestion").WithArgs(qid).WillReturnResult(sqlmock.NewResult(1, 1))
			//批次插入相似問 WithArgs(qid, "相似问题2", "userX", qid, "相似问题3", "userX", qid, "相似问题4", "userX")
			mockedMainDB.ExpectPrepare("INSERT INTO vipshop_squestion").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
			mockedMainDB.ExpectExec("UPDATE ").WillReturnResult(sqlmock.NewResult(1, 1))
			mockedMainDB.ExpectCommit()
			mockedAuditDB.ExpectExec("insert audit_record").WithArgs("userX", "0.0.0.0", util.AuditModuleQA, tt.ExpectedAuditOperation, tt.ExpectedAuditLog, 1).WillReturnResult(sqlmock.NewResult(1, 1))
			e.POST("/question/{qid}/similar-questions").WithPath("qid", qid).WithJSON(SimilarQuestionReqBody{tt.Body}).WithHeaders(mHeader).Expect().Status(http.StatusOK)
			if err := mockedMainDB.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
			if errorLog := errBuff.String(); errorLog != "" {
				t.Fatalf("error have been logged, got %s", errorLog)
			}
		})
	}

}

func TestGetCategoryFullPath(t *testing.T) {
	rows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(1, "LEVEL1", 0).AddRow(2, "LEVEL2", 1).AddRow(3, "LEVEL3", 2)
	mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(rows)
	name, err := GetCategoryFullPath(3)
	if err != nil {
		t.Fatal(err)
	}
	expectedPath := "/LEVEL1/LEVEL2/LEVEL3"
	if name != expectedPath {
		t.Fatalf("expected %s, but got %s", expectedPath, name)
	}

	rows = sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(1, "LEVEL1", 0).AddRow(4, "LEVEL2", 1).AddRow(3, "LEVEL3", 2)
	mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(rows)
	_, err = GetCategoryFullPath(3)
	if err == nil || !strings.Contains(err.Error(), "invalid parentID") {
		t.Fatalf("expected error invalid parentID, but got %+v", err)
	}
}

func TestSelectQuestion(t *testing.T) {
	var qid = 1
	rows := sqlmock.NewRows([]string{"Content", "Category_Id"}).AddRow("Test", 1)
	mockedMainDB.ExpectQuery("SELECT Content, Category_Id from vipshop_question").WillReturnRows(rows)
	mockedMainDB.ExpectQuery("SELECT Content, Category_Id from vipshop_question").WillReturnError(sql.ErrNoRows)
	stdQ, err := selectQuestion(qid, "vipshop")
	if err != nil {
		t.Fatal(err)
	}
	if stdQ.Content != "Test" {
		t.Fatalf("std Question should be Test, but got %s", stdQ.Content)
	}
	if stdQ.CategoryID != 1 {
		t.Fatalf("std Question CategoryId should be 1, but got %d", stdQ.CategoryID)
	}
	stdQ, err = selectQuestion(-1, "vipshop")
	if err != sql.ErrNoRows {
		t.Fatalf("Should return sql.ErrNoRows, but got %v", err)
	}

}

