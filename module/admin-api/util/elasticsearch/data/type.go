package data

type QACoreDoc struct {
	DocID           string       `json:"doc_id"`
	AppID           string       `json:"app_id"`
	Module          string       `json:"module"`
	Answers         []*Answer    `json:"answers,omitempty"`
	Sentence        string       `json:"sentence"`
	SentenceOrig    string       `json:"sentence_original"`
	SentenceType    string       `json:"sentence_type,omitempty"`
	Keywords        string       `json:"keywords,omitempty"`
	Emotions        []*Emotion   `json:"emotions,omitempty"`
	Topics          []*Topic     `json:"topics,omitempty"`
	SpeechActs      []*SpeechAct `json:"speech_acts,omitempty"`
	AutofillEnabled bool         `json:"autofill_enabled"`
}

type Answer struct {
	Sentence   string       `json:"sentence"`
	Emotions   []*Emotion   `json:"emotions,omitempty"`
	Topics     []*Topic     `json:"topics,omitempty"`
	SpeechActs []*SpeechAct `json:"speech_acts,omitempty"`
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
