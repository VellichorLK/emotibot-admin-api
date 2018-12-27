package QA

import (
	"fmt"
	"strings"
)

type QATestInput struct {
	QuestionType string            `json:"qtype"`
	Top          int               `json:"top"`
	Platform     string            `json:"platform"`
	Brand        string            `json:"brand"`
	Gender       string            `json:"sex"`
	Age          string            `json:"age"`
	Hobbies      string            `json:"hobbies"`
	UserInput    string            `json:"text"`
	Info         map[string]string `json:"info"`
}

type RetData struct {
	Emotion         string          `json:"emotion"`
	Intent          string          `json:"intent"`
	SimilarQuestion []*QuestionInfo `json:"similar_question"`
	OpenAPIReturn   int             `json:"openapi_return"`
	Answers         []*string       `json:"answers"`
	Tokens          []*string       `json:"tokens"`
	Module          string          `json:"module"`
}

type QuestionInfo struct {
	UserQuestion     string  `json:"user_q"`
	StandardQuestion string  `json:"std_q"`
	SimilarityScore  float64 `json:"score"`
}

type OpenAPIResponse struct {
	Return       int                    `json:"return"`
	ReturnMsg    string                 `json:"return_message"`
	Status       int                    `json:"status"`
	Data         []TextData             `json:"data"`
	Emotion      []CUData               `json:"emotion"`
	Intent       []CUData               `json:"intent"`
	CustomReturn map[string]interface{} `json:"customReturn"`
}

type TextData struct {
	Type    string        `json:"type"`
	Command string        `json:"cmd"`
	Value   string        `json:"value"`
	Data    []interface{} `json:"data"`
}

type CUData struct {
	Type     string        `json:"type"`
	Score    interface{}   `json:"score"`
	Category string        `json:"category"`
	Data     []interface{} `json:"data"`
	Value    string        `json:"value"`
}

type CUDataFromDC struct {
	Type     string      `json:"type"`
	Score    interface{} `json:"score"`
	Category string      `json:"category"`
	Data     interface{} `json:"data"`
	Item     string      `json:"item"`
}

type DCResponse struct {
	Return  string      `json:"answer"`
	Emotion interface{} `json:"emotion_openapi"`
	Intent  interface{} `json:"intent_openapi"`
	// Emotion []CUDataFromDC `json:"emotion_openapi"`
	// Intent  []CUDataFromDC `json:"intent_openapi"`

	CustomReturn map[string]interface{} `json:"customReturn"`
}

type ControllerScoreRet struct {
	Question string  `json:"question"`
	Score    float64 `json:"score"`
}

type ControllerResponse struct {
	Question        string               `json:"question"`
	Answer          []interface{}        `json:"data"`
	Status          int                  `json:"status"`
	Emotion         string               `json:"emotion"`
	Intent          string               `json:"intent"`
	RelatedQuestion []ControllerScoreRet `json:"relatedQuestions"`
	Tokens          []*string            `json:"tokens"`
}

type InfoNode struct {
	Module       string    `json:"module"`
	Score        float32   `json:"textScore"`
	Intent       string    `json:"intent"`
	IntentScore  float32   `json:"intentScore"`
	Emotion      string    `json:"emotion"`
	EmotionScore float32   `json:"emotionScore"`
	Tokens       []*string `json:"tokens"`
}

type BFOPControllerResponse struct {
	Status int           `json:"status"`
	Answer []interface{} `json:"data"`
	Info   *InfoNode     `json:"info"`
}

type BFOPOpenapiAnswer struct {
	Type    string   `json:"type"`
	SubType string   `json:"subType"`
	Value   string   `json:"value"`
	Data    []string `json:"data"`
}

func (a BFOPOpenapiAnswer) ToString() string {
	if a.Data == nil || len(a.Data) <= 0 {
		return a.Value
	}
	buf := strings.Builder{}
	buf.WriteString(a.Value)
	buf.WriteString("\n")
	for idx := range a.Data {
		buf.WriteString(fmt.Sprintf("%d. %s\n", idx+1, a.Data[idx]))
	}
	return buf.String()
}

type BFOPOpenapiInfoNode struct {
	Module        string    `json:"module"`
	Source        string    `json:"source"`
	Score         float32   `json:"textScore"`
	Emotion       string    `json:"emotion"`
	EmotionScore  float32   `json:"emotionScore"`
	Intent        string    `json:"intent"`
	IntentScore   float32   `json:"intentScore"`
	Tokens        []*string `json:"tokens"`
	MatchQuestion string    `json:"matchQuestion"`
}

type BFOPOpenAPIResponse struct {
	Status     int                  `json:"status"`
	Message    string               `json:"message"`
	Data       []*BFOPOpenapiAnswer `json:"data"`
	Info       *BFOPOpenapiInfoNode `json:"info"`
	CustomInfo interface{}          `json:"customInfo"`
	ExtendData interface{}          `json:"extendData"`
}
