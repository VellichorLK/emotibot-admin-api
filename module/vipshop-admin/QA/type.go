package QA

type QATestInput struct {
	QuestionType string `json:"qtype"`
	Top          int    `json:"top"`
	Platform     string `json:"platform"`
	Brand        string `json:"brand"`
	Gender       string `json:"sex"`
	Age          string `json:"age"`
	Hobbies      string `json:"hobbies"`
	UserInput    string `json:"text"`
}

type RetData struct {
	Emotion         string          `json:"emotion"`
	Intent          string          `json:"intent"`
	SimilarQuestion []*QuestionInfo `json:"similar_question"`
	OpenAPIReturn   int             `json:"openapi_return"`
	Answer          string          `json:"answer"`
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
	Value    string        `json:"value"`
	Score    int           `json:"score"`
	Category string        `json:"category"`
	Data     []interface{} `json:"data"`
}
