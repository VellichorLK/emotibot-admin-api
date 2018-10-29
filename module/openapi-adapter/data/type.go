package data

//ResponseV1 represent version 1 houta response
type ResponseV1 struct {
	ReturnCode int       `json:"return"`
	Message    string    `json:"return_message"`
	Data       []DataV1  `json:"data"`
	Emotion    []Emotion `json:"emotion"`
}

//Data represent version 1 Answer response structure
type DataV1 struct {
	Type  string   `json:"type"`
	Cmd   string   `json:"cmd"`
	Value string   `json:"value"`
	Data  []Answer `json:"data"`
}

type ResponseV2 struct {
	Code    int      `json:"status"`
	Message string   `json:"message"`
	Answers []Answer `json:"data"`
	Info    Info     `json:"info"`
}

type Answer struct {
	Type       string        `json:"type"`
	SubType    string        `json:"subType"`
	Value      string        `json:"value"`
	Data       []interface{} `json:"data"`
	ExtendData string        `json:"extendData"`
}

//Emotion represent version 1 emotion response structure
type Emotion struct {
	Type  string      `json:"type"`
	Value string      `json:"value"`
	Score string      `json:"score"`
	Data  interface{} `json:"data"`
}

type Info struct {
	EmotionCat   string `json:"emotion"`
	EmotionScore int    `json:"emotionScore"`
}

type V2Body struct {
	Text       string                 `json:"text"`
	SourceID   string                 `json:"sourceId,omitempty"`
	ClientID   string                 `json:"clientId,omitempty"`
	CustomInfo map[string]string      `json:"customInfo,omitempty"`
	ExtendData map[string]interface{} `json:"extendData,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
