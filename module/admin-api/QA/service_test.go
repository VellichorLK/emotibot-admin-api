package QA

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"emotibot.com/emotigo/module/admin-api/FAQ"
)

func TestParseDatetimeStr(t *testing.T) {
	var source string = "1999-02-02 12:00:00"
	result, err := parseDatetimeStr(source)
	var target []string = []string{"1999", "02", "02", "12", "00"}

	assert.Equal(t, err, nil, "shold not have error")
	for i := 0; i < 5; i++ {
		assert.Equal(t, result[i], target[i], "should parse correctly")
	}
}

func TestParseKeywordType(t *testing.T) {
	var mockCondition FAQ.QueryCondition = FAQ.QueryCondition{SearchAll: true}
	result := parseKeywordType(&mockCondition)
	target := ""

	assert.Equal(t, result, target, "should parse correctly")
}

func TestGenDimensionStr(t *testing.T) {
	var mockCondition FAQ.QueryCondition = FAQ.QueryCondition{
		Dimension: []FAQ.DimensionGroup{
			FAQ.DimensionGroup{
				TypeId:  1,
				Content: "#aaa#",
			},
			FAQ.DimensionGroup{
				TypeId:  2,
				Content: "#bbb#",
			},
		},
	}
	result := genDimensionStr(&mockCondition)
	target := "aaa,bbb"

	assert.Equal(t, result, target, "should be same")
}

func TestGenCategroyStr(t *testing.T) {
	rows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).
		AddRow(1, "aaa", 0).
		AddRow(2, "bbb", 1).
		AddRow(3, "ccc", 2)

	mainDBMock.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(rows)

	var mockCondition FAQ.QueryCondition = FAQ.QueryCondition{
		CategoryId: 3,
	}
	result, err := genCategoryStr(&mockCondition)
	target := "aaa/bbb"

	assert.Equal(t, err, nil, "should not have error")
	assert.Equal(t, result, target, "should be same")
}

func TestGenQAExportAuditLog(t *testing.T) {
	var mockCondition FAQ.QueryCondition = FAQ.QueryCondition{
		CategoryId: 3,
		Dimension: []FAQ.DimensionGroup{
			FAQ.DimensionGroup{
				TypeId:  1,
				Content: "#aaa#",
			},
			FAQ.DimensionGroup{
				TypeId:  2,
				Content: "#bbb#",
			},
		},
		SearchAll: true,
		Keyword:   "哈哈",
		TimeSet:   true,
		BeginTime: "2011-12-01 12:35:00",
		EndTime:   "2012-12-01 12:35:00",
	}
	var mockTaskID int = 5

	rows := sqlmock.NewRows([]string{"CategoryId", "CategoryName", "ParentId"}).
		AddRow(1, "aaa", 0).
		AddRow(2, "bbb", 1).
		AddRow(3, "ccc", 2)

	mainDBMock.ExpectQuery("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories").WillReturnRows(rows)

	result, err := genQAExportAuditLog(&mockCondition, mockTaskID)
	target := "[部分导出]：[时间段：201112011235-201212011235][全部：哈哈][维度：aaa,bbb][aaa/bbb]：other_5.xlsx"

	assert.Equal(t, err, nil, "should not have error")
	assert.Equal(t, result, target, "should be same")
}
