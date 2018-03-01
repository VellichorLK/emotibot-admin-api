package FAQ

import (
	"database/sql/driver"
	"reflect"
	"strings"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCategoryFullName(t *testing.T) {
	var c = Category{
		ID:       3,
		Name:     "LEVEL3",
		ParentID: 2,
	}
	rows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(1, "LEVEL1", 0).AddRow(2, "LEVEL2", 1).AddRow(3, "LEVEL3", 2)
	mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(rows)
	name, err := c.FullName()
	if err != nil {
		t.Fatal(err)
	}
	expectedPath := "/LEVEL1/LEVEL2/LEVEL3"
	if name != expectedPath {
		t.Fatalf("expected %s, but got %s", expectedPath, name)
	}

	rows = sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).AddRow(1, "LEVEL1", 0).AddRow(4, "LEVEL2", 1).AddRow(3, "LEVEL3", 2)
	mockedMainDB.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(rows)
	_, err = c.FullName()
	if err == nil || !strings.Contains(err.Error(), "invalid parentID") {
		t.Fatalf("expected error invalid parentID, but got %+v", err)
	}
}

func TestSelectQuestions(t *testing.T) {
	type testCase struct {
		input    []int
		expected []StdQuestion
	}
	testCases := map[string]testCase{
		"複數": testCase{
			[]int{1, 2},
			[]StdQuestion{
				StdQuestion{1, "測試標準問A", 1},
				StdQuestion{2, "測試標準問B", 1},
			},
		},
		"找不到": testCase{
			[]int{-1},
			[]StdQuestion{},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			expectSelectQuestions(mockedMainDB, tt.input, tt.expected)
			questions, err := selectQuestions(tt.input, "vipshop")
			if err != nil {
				t.Fatal(err)
			}
			if size := len(questions); size != len(tt.expected) {
				t.Fatalf("select expect size of 1 but got %d", size)
			}
			for i, q := range tt.expected {
				stdQ := questions[i]
				if !reflect.DeepEqual(stdQ, q) {
					t.Fatalf("std Question should be %+v, but got %+v", q, stdQ)
				}
			}

		})
	}
}

func expectSelectQuestions(db sqlmock.Sqlmock, input []int, questions []StdQuestion) {
	rows := sqlmock.NewRows([]string{"Question_id", "Content", "Category_Id"})
	for _, q := range questions {
		rows.AddRow(q.QuestionID, q.Content, q.CategoryID)
	}
	var args = make([]driver.Value, len(input))
	for i, arg := range input {
		args[i] = arg
	}
	db.ExpectQuery("SELECT Question_id, Content, CategoryId from vipshop_question").WithArgs(args...).WillReturnRows(rows)
}
