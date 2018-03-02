package FAQ

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type paramMock struct {
	mock.Mock
}

func (m *paramMock) FormValue(key string) string {
	switch key {
	case "set_time_range":
		return "false"
	case "begin_time":
		return "1970-01-01 00:00:00"
	case "end_time":
		return "2999-12-31 23:59:00"
	case "key_word_type":
		return "relative_question"
	case "key_word":
		return "test"
	case "search_rq":
		return "true"
	case "search_dm":
		return "false"
	case "not_show":
		return "false"
	case "category_id":
		return "0"
	case "page_limit":
		return "10"
	case "cur_page":
		return "0"
	case "dimension":
		return "[]"
	case "search_question":
		return "false"
	case "search_answer":
		return "true"
	case "search_all":
		return "false"
	default:
		return ""
	}
}

func TestFilterQuestion(t *testing.T) {
	mockParam := new(paramMock)
	condition, err := ParseCondition(mockParam)

	targetCondition := QueryCondition{
		TimeSet:                false,
		BeginTime:              "1970-01-01 00:00:00",
		EndTime:                "2999-12-31 23:59:00",
		Keyword:                "test",
		SearchAnswer:           true,
		SearchQuestion:         false,
		SearchDynamicMenu:      false,
		SearchAll:              false,
		SearchRelativeQuestion: true,
		NotShow:                false,
		Dimension:              make([]DimensionGroup, 0),
		CategoryId:             0,
		Limit:                  10,
		CurPage:                0,
	}

	assert.Equal(t, err, nil, "shold not have error")
	assert.Equal(t, condition.TimeSet, targetCondition.TimeSet, "TimeSet should be false")
	assert.Equal(t, condition.BeginTime, targetCondition.BeginTime, "BeginTime should be false")
	assert.Equal(t, condition.EndTime, targetCondition.EndTime, "EndTime should be false")
	assert.Equal(t, condition.Keyword, targetCondition.Keyword, "Keyword should be false")
	assert.Equal(t, condition.SearchRelativeQuestion, targetCondition.SearchRelativeQuestion, "SearchAll should be false")
	assert.Equal(t, condition.SearchDynamicMenu, targetCondition.SearchDynamicMenu, "DynamicMenu should be false")
	assert.Equal(t, condition.SearchRelativeQuestion, targetCondition.SearchRelativeQuestion, "RelativeQuestion should be false")
	assert.Equal(t, condition.SearchQuestion, targetCondition.SearchQuestion, "question should be false")
	assert.Equal(t, condition.SearchAnswer, targetCondition.SearchAnswer, "answer should be false")
	assert.Equal(t, condition.NotShow, targetCondition.NotShow, "NotShow should be false")
	assert.Equal(t, len(condition.Dimension), len(targetCondition.Dimension), "TimeSet should be false")
	assert.Equal(t, condition.CategoryId, targetCondition.CategoryId, "CategoryId should be false")
	assert.Equal(t, condition.Limit, targetCondition.Limit, "Limit should be false")
	assert.Equal(t, condition.CurPage, targetCondition.CurPage, "CurPage should be false")
}
