package cu

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"
)

var isIntegration bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegration, "integration", false, "flag for running integration test")
	flag.Parse()
	os.Exit(m.Run())
}

func TestHandleTextProcessError(t *testing.T) {
	data := []byte(`[{"text": "不開心"}]`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/text/process", bytes.NewBuffer(data))
	handleTextProcess(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Error("expect status code to be 500, but got ", w.Code)
	}
	count := 0
	emotionPredict = func(request emotionengine.PredictRequest) (predictions []emotionengine.Predict, err error) {
		count++
		if count > 1 {
			return nil, fmt.Errorf("testing error is throwed")
		}
		return []emotionengine.Predict{
			emotionengine.Predict{
				Label: "不高興",
				Score: 99,
			},
		}, nil
	}
	data = []byte(`[{"text": "不開心"},{"text": "不爽"}]`)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/text/process", bytes.NewBuffer(data))
	handleTextProcess(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Error("expect status code to be 500, but got ", w.Code)
	}
	if count != 2 {
		t.Error("expect predict to be call 2 time, but got ", count)
	}
}
func TestHandleTextProcess(t *testing.T) {
	emotionPredict = func(request emotionengine.PredictRequest) (predictions []emotionengine.Predict, err error) {
		return []emotionengine.Predict{
			emotionengine.Predict{
				Label: "不高興",
				Score: 99,
			},
			emotionengine.Predict{
				Label: "開心",
				Score: 10,
			},
		}, nil
	}
	data := []byte(`[{"text": "不開心"}]`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/text/process", bytes.NewBuffer(data))
	handleTextProcess(w, r)
	if w.Code != 200 {
		t.Fatal("expect status code to be OK, but got ", w.Code)
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal("reading body failed ")
	}
	var respBody processResponseBody
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		t.Fatal("unmarshal body failed", err)
	}
	goldenResponse := processResponseBody{
		processedText{
			Text: "不開心",
			Emotion: []emotionData{
				emotionData{
					Label: "不高興",
				},
			},
		},
	}
	if !reflect.DeepEqual(respBody, goldenResponse) {
		t.Errorf("result should be %+v\nbut got %+v\n", goldenResponse, respBody)
	}
}
