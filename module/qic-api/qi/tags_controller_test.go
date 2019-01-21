package qi

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestHandleGetTag(t *testing.T) {
	type args struct {
		tagID string
		t     model.Tag
	}
	tests := []struct {
		name     string
		args     args
		hasError bool
		response tag
		mocker   func(t model.Tag) *sql.DB
	}{
		{
			name: "normal case",
			args: args{
				tagID: "test123",
				t: model.Tag{
					ID:               1,
					UUID:             "d218887bdd0e4a979ff0fbe792040907",
					Enterprise:       "47e443e2-4eac-45f7-98d4-38e94175a2a6",
					Name:             "Dexter",
					Typ:              0,
					PositiveSentence: `["test", "test2"]`,
					NegativeSentence: ``,
					IsDeleted:        false,
					CreateTime:       1548059418,
					UpdateTime:       1548059418,
				},
			},
			mocker: func(t model.Tag) *sql.DB {
				db, m, _ := sqlmock.New()
				m.ExpectQuery("SHOW").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("Tag"))
				rows := sqlmock.NewRows([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}).AddRow(
					t.ID, t.UUID, t.IsDeleted,
					t.Name, t.Typ, t.PositiveSentence,
					t.NegativeSentence, t.CreateTime, t.UpdateTime,
					t.Enterprise)
				m.ExpectQuery("SELECT ").WillReturnRows(rows)
				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.mocker(tt.args.t)
			tagDao, _ = model.NewTagSQLDao(db)
			w := httptest.NewRecorder()
			r := mux.SetURLVars(httptest.NewRequest("GET", "/", nil), map[string]string{
				"tag_id": tt.args.tagID,
			})
			HandleGetTag(w, r)
			if !tt.hasError && w.Code != http.StatusOK {
				msg, _ := ioutil.ReadAll(w.Body)
				t.Logf("%s", msg)
				t.Fatal("expect test case ok, but got status ", w.Code)
			}
			data, _ := ioutil.ReadAll(w.Body)
			var resp tag
			err := json.Unmarshal(data, &resp)
			if err != nil {
				t.Fatal("expect handler response to be a valid json, but got error: ", err)
			}
			assertTag(t, tt.args.t, resp)
		})
	}
}
