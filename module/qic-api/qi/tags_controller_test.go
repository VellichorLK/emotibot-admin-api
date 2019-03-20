package qi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"github.com/gorilla/mux"
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
					NegativeSentence: `["test3"]`,
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

var increUpdate200Body = []byte(`[
  {
    "tag_id": "1fe51fd387b14ecda8c2808b5c93dbad",
    "op": "del",
    "sentences": [
      "testFrom範例1"
    ]
  },
  {
    "tag_id": "0043d49adf324e509b23a473c90bfda3",
    "op": "add",
    "sentences": [
      "test2",
      "testFrom範例1"
    ]
  }
]`)

func TestHandleIncreUpdateTagSentences(t *testing.T) {
	type args struct {
		body []byte
	}
	type mock struct {
		Tags      []model.Tag
		updateErr error
	}
	tests := []struct {
		name       string
		args       args
		mock       mock
		wantStatus int
	}{
		{
			name: "200",
			args: args{
				body: increUpdate200Body,
			},
			mock: mock{
				Tags: []model.Tag{
					{
						UUID:             "1fe51fd387b14ecda8c2808b5c93dbad",
						PositiveSentence: `["testFrom範例1","hello"]`,
					},
					{
						UUID:             "0043d49adf324e509b23a473c90bfda3",
						PositiveSentence: `[]`,
					},
				},
			},
			wantStatus: 200,
		},
		{
			name: "DBError",
			args: args{
				body: increUpdate200Body,
			},
			mock: mock{
				Tags: []model.Tag{
					{
						UUID:             "1fe51fd387b14ecda8c2808b5c93dbad",
						PositiveSentence: `["testFrom範例1","hello"]`,
					},
					{
						UUID:             "0043d49adf324e509b23a473c90bfda3",
						PositiveSentence: `[]`,
					},
				},
				updateErr: fmt.Errorf("db is dead"),
			},
			wantStatus: 500,
		},
		{
			name: "invalid json format",
			args: args{
				body: []byte("not even a json"),
			},
			wantStatus: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPut, "http://v1/qi/tags", bytes.NewBuffer(tt.args.body))
			r.Header.Set(requestheader.ConstEnterpriseIDHeaderKey, "csbot")
			tags = func(tx model.SqlLike, query model.TagQuery) ([]model.Tag, error) {
				return tt.mock.Tags, nil
			}
			updateTags = func(tags []model.Tag) error {
				return tt.mock.updateErr
			}
			HandleIncreUpdateTagSentences(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
			t.Log("body: ", w.Body.String())
		})
	}
}
