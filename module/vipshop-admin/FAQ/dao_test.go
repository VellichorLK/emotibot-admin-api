package FAQ

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	expectedPath := "LEVEL1/LEVEL2/LEVEL3"
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
		"sql.ErrNoRows": testCase{
			[]int{-1},
			[]StdQuestion{},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			expectSelectQuestions(mockedMainDB, tt.input, tt.expected)
			questions, err := selectQuestions(tt.input, "vipshop")
			if len(tt.expected) == 0 && err != sql.ErrNoRows {
				t.Errorf("if expect is zero, then return should be sql.ErrNoRows, but got %v", err)
			}
			if len(tt.expected) != 0 && err != nil {
				t.Fatalf("select question err, %v", err)
			}
			if size := len(questions); size != len(tt.expected) {
				t.Fatalf("select expect size of %d but got %d", len(tt.expected), size)
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

func TestEscape(t *testing.T) {
	msgTemplate := "should escape image tag, but get: %s"
	source1 := "<img src=\"some_img_url.jpg\">"
	target1 := "[图片]"
	result1 := Escape(source1)

	msg1 := fmt.Sprintf(msgTemplate, result1)
	assert.Equal(t, result1, target1, msg1)

	source2 := "this string does not need escaping"
	target2 := source2
	result2 := Escape(source2)
	assert.Equal(t, result2, target2, "should not escape")
}

func fakeTagMapFactory() map[string]Tag {
	tagMap := make(map[string]Tag)

	tag1 := Tag{
		Type:    1,
		Content: "tag1",
	}

	tag2 := Tag{
		Type:    2,
		Content: "tag2",
	}

	tag3 := Tag{
		Type:    4,
		Content: "tag3",
	}

	tagMap["1"] = tag1
	tagMap["2"] = tag2
	tagMap["3"] = tag3
	return tagMap
}

func TestFormDimension(t *testing.T) {
	source := []string{"1", "2", "3"}
	target := []string{"tag1", "tag2", "", "tag3", ""}
	tagMap := fakeTagMapFactory()
	result := FormDimension(source, tagMap)

	for i := 0; i < 5; i++ {
		assert.Equal(t, target[0], result[0], "should be same")
	}
}
