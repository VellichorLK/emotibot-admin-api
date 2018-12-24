package emotionengine

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

var isIntegration bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegration, "integrate", false, "")
	flag.Parse()
	os.Exit(m.Run())
}

// TestClientIntegration is a integration test case that run against a local emotion-engine module.
// setup can be found in ./testdata/run.sh
// Scenario:
// 	1. train from a csv file.
//	2. verify by predict api.
func TestClientIntegration(t *testing.T) {
	if !isIntegration {
		t.Skip("Specify -integrate flag for running integration test")
	}

	c := Client{
		Transport: http.DefaultClient,
		ServerURL: "http://localhost:8888",
	}
	e, err := CSVToEmotion("./testdata/happy.csv", "高興")
	if err != nil {
		t.Error("open test data failed, ", err)
	}
	_, err = c.Train(Model{
		AppID:        "demo",
		IsAutoReload: true,
		Data: map[string]Emotion{
			"高興": e,
		},
	})
	if err != nil {
		t.Fatal("expect train to be OK, but got ", err)
	}
	req := PredictRequest{
		AppID:    "demo",
		Sentence: "哈哈，和你开玩笑呢",
	}
	predictions, err := c.Predict(req)
	if err != nil {
		t.Fatal("expect predict to be OK, but got ", err)
	}
	if len(predictions) == 0 {
		t.Fatal("expect at least one prediction but got 0")
	}
	if predictions[0].Label != "高興" {
		t.Error("expect first one to be 高興, but got ", predictions[0].Label)
	}

}

type mockTransporter [][]byte

func (mt mockTransporter) Do(*http.Request) (*http.Response, error) {
	if len(mt) == 0 {
		return nil, fmt.Errorf("you call too many time mock data")
	}
	resp := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(mt[len(mt)-1])),
	}
	mt = mt[:len(mt)-1]
	return resp, nil
}

const exampleTrainRequest = `
{
"status": "OK",
"model_id": "5a3a101dc812670fa745073d"
}`

const examplePredictResult = `
{
"status": "OK", "predictions": [{"label": "不满", "other_info": [], "score": 95}]
}
`

func TestClientTrain(t *testing.T) {
	var transporter = &mockTransporter{
		[]byte(exampleTrainRequest),
	}
	c := Client{
		Transport: transporter,
		ServerURL: "",
	}
	var m = Model{
		AppID:        "csbot",
		IsAutoReload: true,
		Data: map[string]Emotion{
			"不滿": Emotion{
				Name:             "不滿",
				PositiveSentence: []string{"不爽", "你怎麼當客服的"},
				NegativeSentence: []string{},
			},
		},
	}
	id, err := c.Train(m)
	if err != nil {
		t.Fatal("expect OK query but got ", err)
	}
	if id != "5a3a101dc812670fa745073d" {
		t.Fatal("expect id to be 5a3a101dc812670fa745073d, but got ", id)
	}

	c.Transport = nil
	e := m.Data["不滿"]
	e.Name = "亂換的"
	m.Data["不滿"] = e
	c.Train(m)
}

func TestClientPredict(t *testing.T) {
	mt := mockTransporter{
		[]byte(examplePredictResult),
	}
	c := Client{
		Transport: mt,
		ServerURL: "",
	}
	req := PredictRequest{
		AppID:    "csbot",
		Sentence: "測試",
	}
	predictions, err := c.Predict(req)
	if err != nil {
		t.Fatal("expect predict result to be OK, but got ", err)
	}
	if predictions == nil || len(predictions) != 1 {
		t.Fatal("expect predictions to has one element, but got predictions: ", predictions)
	}
	if predictions[0].Label != "不满" {
		t.Error("expect predition 0's labe is 不满, but got ", predictions[0].Label)
	}
	if predictions[0].Score != 95 {
		t.Error("expect prediction 0's Score to be 95, but got ", predictions[0].Score)
	}

}

const failedPredictBody = `{"status": "error","error": "No model loaded."}`

func TestClientPredictFailed(t *testing.T) {
	mt := mockTransporter{
		[]byte(failedPredictBody),
	}
	c := Client{
		Transport: mt,
		ServerURL: "",
	}
	req := PredictRequest{
		AppID:    "csbot",
		Sentence: "測試",
	}
	_, err := c.Predict(req)
	if err == nil {
		t.Fatal("expect predict result to be error, but got nil")
	}

}
