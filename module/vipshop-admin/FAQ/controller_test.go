package FAQ

import (
	"fmt"
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

func TestMain(m *testing.M) {

	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)

	app = iris.New()
	app.Post("/question/{qid:string}/similar-questions", handleUpdateSimilarQuestions)

	retCode := m.Run()
	os.Exit(retCode)
}

//TODO: Multicustomer mock up neeeded!
func TestUpdateHandler(t *testing.T) {
	e := httptest.New(t, app)
	db, mock, err := sqlmock.New()
	qid := "123"
	if err != nil {
		t.FailNow()
	}
	defer db.Close()

	auditdb, mockAudit, err := sqlmock.New()
	if err != nil {
		t.FailNow()
	}
	defer auditdb.Close()

	util.SetDB("main", db)
	util.SetDB("audit", auditdb)
	qRows := sqlmock.NewRows([]string{"Content"}).AddRow("Apple")
	sRows := sqlmock.NewRows([]string{"Content"}).AddRow("It is a sim q")
	mock.ExpectQuery("SELECT Content").WillReturnRows(qRows)
	mock.ExpectQuery("SELECT Content").WillReturnRows(sRows)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM vipshop_squestion").WithArgs(qid).WillReturnResult(sqlmock.NewResult(1, 1))
	// mock.ExpectExec("")
	mock.ExpectPrepare("INSERT INTO vipshop_squestion").ExpectExec().WithArgs(qid, "123", "userX").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE ").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// mockAudit.ExpectBegin()
	mockAudit.ExpectExec("insert audit_record").WillReturnResult(sqlmock.NewResult(1, 1))
	// mockAudit.ExpectCommit()
	body := SimilarQuestionReqBody{
		SimilarQuestions: []SimilarQuestion{
			SimilarQuestion{
				Content: "123",
				Id:      "1",
			},
		},
	}
	e.POST("/question/{qid}/similar-questions").WithPath("qid", qid).WithJSON(body).WithHeaders(mHeader).Expect().Status(http.StatusOK)

	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Println(body)
		t.Fatal(err.Error)
	}
}
