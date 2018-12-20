package qi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
	"net/http"
	"net/http/httptest"
)


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
	r := httptest.NewRequest(http.MethodGet, "/qi/groups",  bytes.NewBuffer(reqBody))
	handleCreateGroup(w, r)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}

	group := Group{}
	err = json.Unmarshal(body, &group)
	if err != nil {
		t.Error(err)
		return
	}

	if group.ID != 55688 {
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

	groups := []Group{}
	json.Unmarshal(body, &groups)

	if len(groups) != 2 {
		t.Errorf("expect 2 groups but got %d", len(groups))
		return
	}

	for idx := range groups {
		g := groups[idx]
		targetG := mockGroups[idx]

		if g.ID != targetG.ID || g.Name != targetG.Name {
			t.Errorf("expect ID: %d, Name: %s, but got %d, %s", targetG.ID, targetG.Name, g.ID, g.Name)
			return
		}
	}
}