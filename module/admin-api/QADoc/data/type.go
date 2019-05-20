package data

type QACoreDoc struct {
	DocID           string       `json:"doc_id"`
	AppID           string       `json:"app_id"`
	Module          string       `json:"module"`
	Domain          string       `json:"domain"`
	Answers         []*Answer    `json:"answers,omitempty"`
	Sentence        string       `json:"sentence"`
	SentenceOrig    string       `json:"sentence_original"`
	SentenceType    string       `json:"sentence_type,omitempty"`
	SentencePos     string       `json:"sentence_pos,omitempty"`
	Keywords        string       `json:"keywords,omitempty"`
	Source          string       `json:"source,omitempty"`
	Emotions        []*Emotion   `json:"emotions,omitempty"`
	Topics          []*Topic     `json:"topics,omitempty"`
	SpeechActs      []*SpeechAct `json:"speech_acts,omitempty"`
	AutofillEnabled bool         `json:"autofill_enabled"`
	StdQID          string       `json:"std_q_id"`
	StdQContent     string       `json:"std_q_content"`
}

type Answer struct {
	Sentence   string       `json:"sentence"`
	Emotions   []*Emotion   `json:"emotions,omitempty"`
	Topics     []*Topic     `json:"topics,omitempty"`
	SpeechActs []*SpeechAct `json:"speech_acts,omitempty"`
	Creator    string       `json:"creator,omitempty"`
}

type Emotion struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type Topic struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type SpeechAct struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type AutofillToggleBody struct {
	ID      string `json:"autofill_id"`
	Enabled bool   `json:"autofill_enabled"`
}

type RegexQuery struct {
	Field      string
	Expression string
}

type DeleteQADocsRequest struct {
	AppID  string `json:"app_id"`
	Module string `json:"module"`
}

type DeleteQADocsByIDsRequest struct {
	IDs []interface{} `json:"ids"`
}
