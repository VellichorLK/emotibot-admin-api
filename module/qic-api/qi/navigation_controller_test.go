package qi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestDetailFlowToSetting(t *testing.T) {
	d := &DetailNavFlow{}
	d.IntentLinkID = 1
	d.Role = 1
	d.IgnoreIntent = 0
	d.IntentName = "hello world"
	r := detailFlowToSetting(d)
	if r == nil {
		t.Fatalf("expecintg r is not nil, but get nil")
	}

	if callInIntentCodeMap[1] != r.Type {
		t.Fatalf("expecting %s type, but get %s\n", callInIntentCodeMap[d.IgnoreIntent], r.Type)
	}

	if roleCodeMap[d.Role] != r.Role {
		t.Fatalf("expecting %s role, but get %s\n", roleCodeMap[d.Role], r.Role)
	}

	if d.IntentName != r.IntentName {
		t.Fatalf("expecting %s intent name, but get %s\n", d.IntentName, r.IntentName)
	}
}

func TestHandleGetFlowSetting(t *testing.T) {
	setUpMockNav()
	uuid := "sss"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/qi/call-in/navigation/"+uuid, nil)
	r = mux.SetURLVars(r, map[string]string{"id": uuid})
	r.Header.Add("X-EnterpriseID", mockEnterprise)
	handleGetFlowSetting(w, r)

	if http.StatusBadRequest != w.Code {
		t.Errorf("expecting get %d status, but get %d\n", http.StatusBadRequest, w.Code)
	}
}

func TestHandleNewFlow(t *testing.T) {
	setUpMockNav()

	rb := &reqNewFlow{Name: "hello world"}
	rbb, err := json.Marshal(rb)
	if err != nil {
		t.Error(err)
		return
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/qi/call-in/navigation", bytes.NewBuffer(rbb))
	r.Header.Add("X-EnterpriseID", mockEnterprise)
	handleNewFlow(w, r)

	if http.StatusOK != w.Code {
		t.Fatalf("expecting status %d , but get %d\n", http.StatusOK, w.Code)
	}
}

func TestHandleFlowList(t *testing.T) {

	setUpMockNav()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/qi/call-in/navigation", nil)
	r.Header.Add("X-EnterpriseID", mockEnterprise)
	handleFlowList(w, r)

	if http.StatusOK != w.Code {
		t.Errorf("expecting get %d status, but get %d\n", http.StatusOK, w.Code)
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}

	var resp RespFlowList
	err = json.Unmarshal(body, &resp)
	if err != nil {
		t.Fatalf("Expecting get sentencesResp struct, but get %s. %s\n", body, err)
	}
	if len(mockFlows) != len(resp.Data) {
		t.Fatalf("expecting get %d flows, but get %d\n", len(mockFlows), len(resp.Data))
	}

	for idx, expect := range mockFlows {
		f := resp.Data[idx]
		if expect.ID != f.ID {
			t.Fatalf("expecting get id %d at index %d, but get id %d\n", expect.ID, idx, f.ID)
		}
		if mockCountNodes[expect.ID] != f.NumOfNodes {
			t.Fatalf("expecting id %d get %d nodes, but get %d\n", expect.ID, mockCountNodes[expect.ID], f.NumOfNodes)
		}
		if expect.Name != f.Name {
			t.Fatalf("expecting id %d get flow name %s, but get %s\n", expect.ID, expect.Name, f.Name)
		}
		if expect.IntentName != f.IntentName {
			t.Fatalf("expecting id %d get flow name %s, but get %s\n", expect.ID, expect.IntentName, f.IntentName)
		}
	}
}

func TestHandleDeleteFlow(t *testing.T) {
	setUpMockNav()
	uuid := "1234"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/qi/call-in/navigation/"+uuid, nil)
	r = mux.SetURLVars(r, map[string]string{"id": uuid})
	r.Header.Add("X-EnterpriseID", mockEnterprise)
	handleDeleteFlow(w, r)

	if http.StatusOK != w.Code {
		t.Errorf("expecting get %d status, but get %d\n", http.StatusOK, w.Code)
	}
}

func TestHandleNewNode(t *testing.T) {
	setUpMockNav()

	rb := &SentenceGroupInReq{Name: "hello world", Role: "staff"}
	rbb, err := json.Marshal(rb)
	if err != nil {
		t.Error(err)
		return
	}

	uuid := "1234"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/qi/call-in/navigation/"+uuid+"/node", bytes.NewBuffer(rbb))
	r = mux.SetURLVars(r, map[string]string{"id": uuid})
	r.Header.Add("X-EnterpriseID", mockEnterprise)
	handleNewNode(w, r)

	if http.StatusOK != w.Code {
		t.Errorf("expecting get %d status, but get %d\n", http.StatusOK, w.Code)
	}

}

func TestHandleModifyIntent(t *testing.T) {

	setUpMockNav()
	uuid := "1234"
	rb := &reqCallInIntent{IntentName: "hello world", Role: "staff", Type: "fixed"}
	rbb, err := json.Marshal(rb)
	if err != nil {
		t.Error(err)
		return
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/qi/call-in/navigation/"+uuid+"/intent", bytes.NewBuffer(rbb))
	r = mux.SetURLVars(r, map[string]string{"id": uuid})
	r.Header.Add("X-EnterpriseID", mockEnterprise)
	handleModifyIntent(w, r)

	if http.StatusOK != w.Code {
		t.Errorf("expecting get %d status, but get %d\n", http.StatusOK, w.Code)
	}

}
