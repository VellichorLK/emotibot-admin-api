package qi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
	// mockDAO is defined in service_test.go
	originDAO := serviceDAO
	m := &mockDAO{}
	serviceDAO = m
	defer restoreDAO(originDAO)

	reqBody, err := json.Marshal(mockGroup)
	if err != nil {
		t.Error(err)
		return
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/qi/groups", bytes.NewBuffer(reqBody))
	handleCreateGroup(w, r)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}

	group := model.GroupWCond{}
	err = json.Unmarshal(body, &group)
	if err != nil {
		t.Error(err)
		return
	}

	if group.UUID != "abcde" {
		t.Error("create group failed")
		return
	}
}

func TestHandleCreateGroup(t *testing.T) {
	// mockDAO is defined in service_test.go
	originDAO := serviceDAO
	m := &mockDAO{}
	serviceDAO = m
	defer restoreDAO(originDAO)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/qi/groups", nil)
	handleGetGroups(w, r)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}

	response := SimpleGroupsResponse{}
	json.Unmarshal(body, &response)

	if response.Paging.Total != int64(len(mockGroups)) {
		t.Errorf("expect 2 groups but got %d", response.Paging.Total)
		return
	}

	for idx := range response.Data {
		g := response.Data[idx]
		targetG := mockGroups[idx]

		if g.ID != targetG.UUID || g.Name != targetG.Name {
			t.Errorf("expect ID: %s, Name: %s, but got %s, %s", targetG.UUID, targetG.Name, g.ID, g.Name)
			return
		}
	}
}

func TestHandleGetGroup(t *testing.T) {
	// mockDAO is defined in service_test.go
	originDAO := serviceDAO
	m := &mockDAO{}
	serviceDAO = m
	defer restoreDAO(originDAO)

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/groups/55688", nil)
	if err != nil {
		t.Error(err)
		return
	}

	router := getTestRouter()

	router.ServeHTTP(w, r)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("body: %s\n", body)
	group := model.GroupWCond{}
	err = json.Unmarshal(body, &group)
	if err != nil {
		t.Error(err)
		return
	}

	if !sameGroup(&group, mockGroup) {
		t.Errorf("expect group: %+v, but got %+v", mockGroup, group)
		return
	}
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

	if filter.FileName != values.Get("file_name") || filter.Deal != 1 || filter.Series != values.Get("series") || filter.CallStart != 10056 {
		t.Error("parse group filter failed")
		return
	}
}
