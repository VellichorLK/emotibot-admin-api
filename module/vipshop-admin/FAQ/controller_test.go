package FAQ

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"os"
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

	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	app = iris.New()
	app.Post("/question/{qid:string}/similar-questions", handleUpdateSimilarQuestions)
	app.Get("/RFQuestions", handleGetRFQuestions)
	app.Post("/RFQuestions", HandleSetRFQuestions)
	app.Get("/category/{cid:int}/questions", handleCategoryQuestions)

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
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, errBuff)
	for _, tt := range testCases {
		t.Run(tt.Scenario, func(t *testing.T) {
			errBuff.Reset()
			//設定場景: /LEVEL1/LEVEL2/LEVEL3 之 標準問題1，有相似問題1,4。 更新相似問 2,3,4。
			cRows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(1, "LEVEL1", 0).AddRow(2, "LEVEL2", 1).AddRow(3, "LEVEL3", 2)
			sRows := sqlmock.NewRows([]string{"Content"}).AddRow("相似问题1").AddRow("相似问题4")
			expectSelectQuestions(mockedMainDB, []int{qid}, []StdQuestion{StdQuestion{qid, "标准问题1", 3}})
			//Find Category
			mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(3, "LEVEL3", 2))
			//Category's Full Name
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
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

}

func TestHandleGetRFQuestions(t *testing.T) {
	e := httptest.New(t, app)
	var expected = []RFQuestion{
		RFQuestion{1, "測試A"},
		RFQuestion{2, "測試B"},
	}

	rows := sqlmock.NewRows([]string{"rd.id", " rf.Question_Content"})
	for _, q := range expected {
		rows.AddRow(q.ID, q.Content)
	}
	mockedMainDB.ExpectQuery("SELECT ").WillReturnRows(rows)
	resp := e.GET("/RFQuestions").Expect().Body()
	if mockedMainDB.ExpectationsWereMet() != nil {
		t.Fatal()
	}
	expectedJSON, _ := json.Marshal(expected)
	resp.Equal(string(expectedJSON))
}

func TestHandleSetRFQuestions(t *testing.T) {
	e := httptest.New(t, app)
	input := []int{1, 2, 3}
	expected := []StdQuestion{
		StdQuestion{1, "測試A", 1},
		StdQuestion{2, "測試B", 1},
		StdQuestion{3, "測試C", 1},
	}
	expectSelectQuestions(mockedMainDB, input, expected)
	// var contents = []driver.Value{"測試A", "測試B", "測試C"}
	mockedMainDB.ExpectExec("INSERT INTO vipshop_removeFeedbackQuestion").WillReturnResult(sqlmock.NewResult(4, 4))

	request := e.POST("/RFQuestions").WithHeader("Authorization", "vipshop")
	for _, i := range input {
		request.WithQuery("id", i)
	}
	request.Expect().Status(200)
	if err := mockedMainDB.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetQuestionsByCategoryId(t *testing.T) {
	type input struct {
		categoryID    int
		includeSubCat bool
	}
	type testCase struct {
		input    input
		expected []StdQuestion
	}

	testCases := map[string]testCase{
		"一般": testCase{
			input{
				1,
				true,
			},
			[]StdQuestion{
				StdQuestion{1, "123", 1},
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			e := httptest.New(t, app)
			request := e.GET("/category/{cid:int}/questions", tt.input.categoryID).WithQuery("includeSubCat", true)
			//Select the Category
			rows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(tt.input.categoryID, "Test", 0)
			mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WithArgs(tt.input.categoryID).WillReturnRows(rows)
			//Select SubCategory
			// rows = sqlmock.NewRows([]string{"CategoryId", "CategoryName"})
			// mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName FROM vipshop_categories").WillReturnRows(rows)
			//Select Questions
			rows = sqlmock.NewRows([]string{"Question_id", "Content", "CategoryId"})
			for _, q := range tt.expected {
				rows.AddRow(q.QuestionID, q.Content, q.CategoryID)
			}
			mockedMainDB.ExpectQuery("SELECT Question_id, Content, CategoryId FROM vipshop_question WHERE CategoryId IN ").WillReturnRows(rows)
			response := request.Expect().Status(200).Body()
			expectedResponse, _ := json.Marshal(tt.expected)
			response.Equal(string(expectedResponse))
		})
	}
}

