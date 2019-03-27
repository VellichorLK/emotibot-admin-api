package qi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"github.com/gorilla/mux"
)

func getTestRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	for _, entrypoint := range ModuleInfo.EntryPoints {
		entryPath := fmt.Sprintf("/%s", entrypoint.EntryPath)
		r.Methods(entrypoint.AllowMethod).
			Path(entryPath).
			Name(entrypoint.EntryPath).
			HandlerFunc(entrypoint.Callback)
	}
	return r
}

func TestHandleGetGroups(t *testing.T) {
	expect := `
{
  "paging": {
    "page": 1,
    "total": 1,
    "limit": 10
  },
  "data": [
    {
      "group_id": "45f8577451d0478095197ega3d054133",
			"group_name": "a group",
			"description": "",
      "is_enable": 1,
      "other": {
				"type": 0,
        "file_name": "example.wav",
        "call_time": 78923981273,
        "deal": 1,
        "series": "abcd-123",
        "staff_id": "host-abcd",
        "staff_name": "Melvina",
        "extension": "",
				"department": "",
        "customer_id": "guest-123",
        "customer_name": "Nina",
        "customer_phone": "886-1234-5678",
        "left_channel": "staff",
        "right_channel": "customer",
        "call_from": 78923981273,
        "call_end": 78923981273
      },
      "create_time": 0,
      "rule_count": 56
    }
  ]
}
`
	tmp := groupResps
	defer func() {
		groupResps = tmp
	}()
	groupResps = func(filter *model.GroupFilter) (int64, []GroupResp, error) {
		return 1, []GroupResp{
			GroupResp{
				GroupID:   "45f8577451d0478095197ega3d054133",
				GroupName: "a group",
				IsEnable:  1,
				Other: Other{
					Type:          0,
					FileName:      "example.wav",
					CallTime:      78923981273,
					Deal:          1,
					Series:        "abcd-123",
					StaffID:       "host-abcd",
					StaffName:     "Melvina",
					CustomerID:    "guest-123",
					CustomerName:  "Nina",
					CustomerPhone: "886-1234-5678",
					LeftChannel:   "staff",
					RightChannel:  "customer",
					CallFrom:      78923981273,
					CallEnd:       78923981273,
				},
				CreateTime: 0,
				RuleCount:  56,
			},
		}, nil
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/qi/groups", nil)
	q := r.URL.Query()
	q.Set("limit", "10")
	q.Set("page", "1")
	r.URL.RawQuery = q.Encode()
	handleGetGroups(w, r)
	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, expect, w.Body.String())
}

func TestHandleCreateGroup(t *testing.T) {
	var (
		expect = `{"group_id": "45f8577451d0478095197ega3d054133"}`
	)
	tmp := newGroupWithAllConditions
	tmp2 := getConversationRulesBy
	defer func() {
		newGroupWithAllConditions = tmp
		getConversationRulesBy = tmp2
	}()

	newGroupWithAllConditions = func(group model.Group, condition model.Condition, customCols map[string][]interface{}) (model.Group, error) {
		return model.Group{
			ID:   1,
			UUID: "45f8577451d0478095197ega3d054133",
		}, nil
	}
	getConversationRulesBy = func(filter *model.ConversationRuleFilter) (int64, []model.ConversationRule, error) {
		return 0, []model.ConversationRule{}, nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/qi/groups", strings.NewReader("{}"))
	handleCreateGroup(w, r)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, expect, w.Body.String(), "expect json equal")
}

func TestHandleGetGroup(t *testing.T) {
	// mockDAO is defined in service_test.go
	tmp := getConditionOfGroup
	tmp1 := customConditionsOfGroup
	defer func() {
		getConditionOfGroup = tmp
		customConditionsOfGroup = tmp1
	}()
	getConditionOfGroup = func(groupID int64) (*model.Condition, error) {
		return &model.Condition{}, nil
	}
	customConditionsOfGroup = func(groupID int64) (map[string][]interface{}, error) {
		return make(map[string][]interface{}, 0), nil
	}
	// mg := mockGroup
	// mg.Rules = &[]model.ConversationRule{}
	// mg.RuleCount = 0
	// w := httptest.NewRecorder()
	// r := httptest.NewRequest(http.MethodGet, "http://testing/groups/ABCDE", nil)
	// r.Header.Set(requestheader.ConstEnterpriseIDHeaderKey, "csbot")
	// handleGetGroup(w, r, mockGroup)
	// require.Equal(t, http.StatusOK, w.Code, "Body: %s", w.Body.String())
	// expect := `{"group_id":"123456","group_name":"group_name","is_enable":1,"other":{"call_end":0,"call_from":0,"call_time":0,"customer_id":"","customer_name":"","customer_phone":"","deal":0,"department":"","extension":"","file_name":"","left_channel":"staff","right_channel":"staff","series":"","staff_id":"","staff_name":"","type":0},"create_time":0,"description":"group_description","rule_count":0,"rules":[]}`
	// assert.JSONEq(t, expect, w.Body.String())
}

func TestParseGroupFilter(t *testing.T) {
	values := url.Values{}
	values.Add("file_name", "abcd.wmv")
	values.Add("deal", "1")
	values.Add("series", "test")
	values.Add("call_start", "10056")

	filter, err := parseGroupFilter(&values)
	if err != nil {
		t.Error(err)
		return
	}

	if filter.FileName != values.Get("file_name") || filter.Series != values.Get("series") || filter.CallStart != 10056 {
		t.Error("parse group filter failed")
		return
	}
}
