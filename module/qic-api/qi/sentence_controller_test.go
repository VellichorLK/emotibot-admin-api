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

func TestHandleGetSentences(t *testing.T) {
	sentenceMockDataSetup()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/qi/sentences", nil)
	enterprise := mockSentenceDao.enterprises[0]
	r.Header.Add("X-EnterpriseID", enterprise)
	handleGetSentences(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expecting get status %d, but get %d\n", http.StatusOK, w.Code)
		return
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}
	var resp sentencesResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		t.Fatalf("Expecting get sentencesResp struct, but get %s. %s\n", body, err)
	}
	if len(resp.Data) != mockSentenceDao.numByEnterprise[enterprise] {
		t.Errorf("Expecting get %d records, but get %d, %v\n",
			mockSentenceDao.numByEnterprise[enterprise], len(resp.Data), resp)
	}

}
func mockBuildParameters(v map[string]string) map[string]string {
	return v
}

func TestHandleGetSentence(t *testing.T) {
	sentenceMockDataSetup()
	for _, v := range mockSentenceDao.data {

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/qi/sentences"+v.UUID, nil)
		r = mux.SetURLVars(r, map[string]string{"id": v.UUID})
		r.Header.Add("X-EnterpriseID", v.Enterprise)
		handleGetSentence(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expecting get status %d, but get %d\n", http.StatusOK, w.Code)
			continue
		}

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Error(err)
			continue
		}

		var resp sentenceResp
		err = json.Unmarshal(body, &resp)
		if err != nil {
			t.Errorf("Expecting get sentenceResp struct, but get %s. %s\n", body, err)
			continue
		}
		if resp.UUID != v.UUID {
			t.Errorf("Expecting get UUID %s, but get %s\n", v.UUID, resp.UUID)
		}
	}
}

func TestHandleDeleteSentence(t *testing.T) {

	sentenceMockDataSetup()

	for uuid, v := range mockSentenceDao.uuidData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/qi/sentences"+uuid, nil)
		r = mux.SetURLVars(r, map[string]string{"id": uuid})
		r.Header.Add("X-EnterpriseID", v.Enterprise)

		if w.Code != http.StatusOK {
			t.Errorf("Expecting get status %d, but get %d\n", http.StatusOK, w.Code)
			continue
		}
		handleDeleteSentence(w, r)
		if v.IsDelete != 1 {
			t.Errorf("Expecting get %s isDeleted, but get no isDeleted", uuid)
		}
	}
}

func TestHandleNewSentence(t *testing.T) {
	sentenceMockDataSetup()

	//testing empty body request
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/qi/sentences", nil)
	r.Header.Add("X-EnterpriseID", mockSentenceDao.enterprises[0])
	handleNewSentence(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expecting get status %d, but get %d\n", http.StatusBadRequest, w.Code)
		return
	}

	//testing create new sentence
	rb := sentenceReq{Name: "mymock", Tags: []string{mockTagDao.uuid[0]}}

	rbb, err := json.Marshal(rb)
	if err != nil {
		t.Error(err)
		return
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/qi/sentences", bytes.NewBuffer(rbb))
	r.Header.Add("X-EnterpriseID", mockSentenceDao.enterprises[0])

	handleNewSentence(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expecting get status %d, but get %d\n", http.StatusOK, w.Code)
		return
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}
	var resp sentenceResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		t.Errorf("Expecting get sentenceResp struct, but get %s. %s\n", body, err)
		return
	}
	uuid := resp.UUID
	if v, ok := mockSentenceDao.uuidData[uuid]; ok {
		if uuid != v.UUID {
			t.Errorf("Expecting get data %s in db, but get %s\n", uuid, v.UUID)
		}
	} else {
		t.Errorf("Expecting data %s in db, but get none\n", uuid)
	}

}
func TestHandleUpdateSentence(t *testing.T) {
	sentenceMockDataSetup()
	uuid := mockSentenceDao.uuid[0]
	enterprise := mockSentenceDao.enterprises[0]

	mockName := "testupdate"
	mockTags := []string{mockTagDao.uuid[0]}

	rb := sentenceReq{Name: mockName, Tags: mockTags}
	rbb, err := json.Marshal(rb)
	if err != nil {
		t.Error(err)
		return
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/qi/sentences"+uuid, bytes.NewBuffer(rbb))
	r = mux.SetURLVars(r, map[string]string{"id": uuid})
	r.Header.Add("X-EnterpriseID", enterprise)
	handleModifySentence(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expecting get status %d, but get %d\n", http.StatusOK, w.Code)
		return
	}

	if v, ok := mockSentenceDao.uuidData[uuid]; ok {
		if v.Name != mockName {
			t.Errorf("Excpecting get name %s in db, but get %s\n", mockName, v.Name)
		}
		if len(v.TagIDs) != len(mockTags) {
			t.Errorf("Excpecting get %d tag, but get %d\n", len(mockTags), len(v.TagIDs))
		}

		if tag, ok := mockTagDao.data[v.TagIDs[0]]; ok {
			if tag.UUID != mockTags[0] {
				t.Errorf("Excpecting get tag %s in db, but get %s\n", mockTags[0], tag.UUID)
			}
		} else {
			t.Errorf("Expecting tag %d in db, but get none\n", v.TagIDs[0])
		}

	} else {
		t.Errorf("Expecting data %s in db, but get none\n", uuid)
	}
}
