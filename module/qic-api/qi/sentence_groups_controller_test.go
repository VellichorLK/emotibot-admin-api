package qi

import (
	"bytes"
	_ "emotibot.com/emotigo/module/qic-api/model/v1"
	"encoding/json"
	_ "fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCreateSentenceGroup(t *testing.T) {
	originDBLike, originSGDao, originSDao := setupSentenceGroupTestMock()
	defer restoreSentenceGroupTest(originDBLike, originSGDao, originSDao)

	sg := SentenceGroupInReq{
		Name:             mockSentenceGroup1.Name,
		Role:             "staff",
		Position:         "top",
		PositionDistance: 5,
		Sentences: []string{
			mockSimpleSentence1.UUID,
			mockSimpleSentence2.UUID,
		},
	}

	reqBody, err := json.Marshal(sg)
	if err != nil {
		t.Error(err)
		return
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/qi/sentence-groups", bytes.NewBuffer(reqBody))
	handleCreateSentenceGroup(w, r)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
		return
	}

	group := SentenceGroupInResponse{}
	err = json.Unmarshal(body, &group)
	if err != nil {
		t.Error(err)
		return
	}

	if group.ID != mockSentenceGroup1.UUID {
		t.Errorf("expect sentence group id: %s, but got: %s", mockSentenceGroup1.UUID, group.ID)
		return
	}
}
