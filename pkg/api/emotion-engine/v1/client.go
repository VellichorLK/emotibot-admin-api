//Package emotionengine provide a api client to interact with emotion-engine module
package emotionengine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/pkg/api"
)

//Client is the api client to communicate with
type Client struct {
	// Transport specify how the HTTP Request are handled.
	// Which should be compatible with the net/http package.
	// If nil, http.DefaultClient is used.
	Transport api.HTTPClient
	// ServerURL is the host and port of the emotion-engine module.
	// which shouold be "http://{host}:{port}"
	ServerURL string
}

// Train will tell emotion-engine to create a model based on the model's AppID
// It will return modelID if successful. On error, modelID will be empty
// returned error can be http, emotion-engine, or other error.
// It is seperated by message prefix HTTP, EE, or OTHER.
func (c *Client) Train(apiModel Model) (modelID string, err error) {
	var transporter api.HTTPClient

	if c.Transport == nil {
		transporter = http.DefaultClient
	} else {
		transporter = c.Transport
	}

	m := model{
		AppID:      apiModel.AppID,
		AutoReload: apiModel.IsAutoReload,
		Data:       make(map[string]interface{}, 0),
	}

	emotions := make([]emotion, 0, len(apiModel.Data))
	for name, e := range apiModel.Data {
		if name != e.Name {
			return "", fmt.Errorf("EE: invalid model data, EmotionName '%s' should be the same with the key of it's Data map '%s'", e.Name, name)
		}
		if e.PositiveSentence == nil || len(e.PositiveSentence) == 0 {
			return "", fmt.Errorf("EE: invalid model data, Emotion %s's PositiveSentence should have at least one element", name)
		}
		var negativeSentence []string
		// Handling edge case. If we send null or negative sentence is not presented in json.
		// emotion-engine(last verified version: 58a9eb1) will return error. but the spec said it's optional.
		if e.NegativeSentence == nil {
			negativeSentence = []string{}
		} else {
			negativeSentence = e.NegativeSentence
		}
		emotions = append(emotions, emotion{
			EmotionName: e.Name,
			Sentences: sentence{
				Positive: e.PositiveSentence,
				Negative: negativeSentence,
			},
		})
	}
	m.Data["emotion"] = emotions

	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("OTHER: json marshal failed, %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.ServerURL+"/train", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("OTHER: New Request failed, %v", err)
	}
	resp, err := transporter.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP: %v", err)
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	var respBody trainResponse
	err = json.Unmarshal(data, &respBody)
	if err != nil {
		logger.Error.Println("raw output: ", string(data))
		return "", fmt.Errorf("EE: response format invalid, %v", err)
	}
	if respBody.Status != "OK" {
		return "", fmt.Errorf("EE: got unsuccessful status '%s', error message: %s", respBody.Status, respBody.ErrMsg)
	}
	return respBody.ModelID, nil
}

// Predict will sending request to emotion-engine to get the PredictResult from it.
// It will return a slice of Predicts in response of the request.Sentence.
// returned error can be http, emotion-engine, or other error.
// It is seperated by message prefix HTTP, EE, or OTHER.
func (c *Client) Predict(request PredictRequest) (predictions []Predict, err error) {
	if request.AppID == "" {
		return nil, fmt.Errorf("EE: predict appID should not be empty")
	}
	data, err := json.Marshal(request)
	req, err := http.NewRequest(http.MethodPost, c.ServerURL+"/predict", bytes.NewBuffer(data))
	resp, err := c.Transport.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP: Transport Do request failed, %v", err)
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP: io read failed, %v", err)
	}
	var respBody predictResponse
	err = json.Unmarshal(data, &respBody)
	if err != nil {
		logger.Error.Println("raw output: ", string(data))
		return nil, fmt.Errorf("EE: predict response is not valid, %v", err)
	}

	if respBody.Status != "OK" {
		return nil, fmt.Errorf("EE: got unsuccessful status '%s', error message: %s", respBody.Status, respBody.Error)
	}

	var result = []Predict{}
	for _, p := range respBody.Predictions {
		result = append(result, Predict{
			Label: p.Label,
			Score: p.Score,
		})
	}
	return result, nil
}

// Model contain config and data of a emotion model, which can be used to predict emotion.
// Model also is the input value for train api.
// Be remind, it is not an 1 to 1 mapping of original
type Model struct {
	// AppID will be the unique id of the model.
	// Any two equal AppID model will be override in sequence manner.
	AppID string
	//IsAutoReload is an indicator if the model need to loaded on trained.
	IsAutoReload bool
	//Data is used for training, and the key should be the same of Emotion's Name.
	//All Emotions should not have duplicated name.
	Data map[string]Emotion
}

//Emotion represent an emotion of the model. The name should always be unique in model.
type Emotion struct {
	//name is the key in the emotion model
	Name string
	//PositiveSentence is a required argument. it should never be nil.
	PositiveSentence []string
	//NegativeSentence is an optional argument. it can be nil or empty.
	NegativeSentence []string
}

//PredictRequest is the input of Client.Predict api.
type PredictRequest struct {
	AppID    string `json:"app_id"`
	Sentence string `json:"sentence"`
}

//Predict is the result of Client.Predict().
type Predict struct {
	// Label is the category tag for the input sentence, which should be the same with Emotion Name.
	Label string
	// Score is the confident score of the Label, from 0 to 100.
	// Threshold should be determine by the user.
	Score int
}
type predictResponse struct {
	Status      string       `json:"status"`
	Predictions []rawPredict `json:"predictions"`
	Error       string       `json:"error"`
}

const (
	errPredictNotLoad = "No model loaded."
	errPredictLoading = "Model is loading."
)

var (
	// ErrModelNotLoad is the predefined error from emotion-engine.
	// Which mean model has not issued to load yet. or mode is not found.
	ErrModelNotLoad = errors.New("predifined model not loaded message has been returned")
	// ErrModelLoading is the predfined error from emotion-engine.
	// Which mean model hasn't finish loading yet. please try again later.
	ErrModelLoading = errors.New("")
)

type rawPredict struct {
	Label     string        `json:"label"`
	Score     int           `json:"score"`
	OtherInfo []interface{} `json:"other_info"`
}

type model struct {
	AppID      string                 `json:"app_id"`
	AutoReload bool                   `json:"auto_reload"`
	Data       map[string]interface{} `json:"data"`
}

type emotion struct {
	EmotionName string   `json:"emotion_name"`
	Sentences   sentence `json:"sentences"`
}

type sentence struct {
	Positive []string `json:"positive"`
	Negative []string `json:"negative"`
}

type trainResponse struct {
	Status  string `json:"status"`
	ModelID string `json:"model_id"`
	ErrMsg  string `json:"error"`
}
