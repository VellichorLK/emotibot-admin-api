package data

import "time"

const (
	TaskStatusRunning = iota
	TaskStatusFinished
	TaskStatusExpired
	TaskStatusError
)

type AutofillOption struct {
	Module   string
	TaskMode int64
}

type AutofillBody struct {
	ModuleID         int64  `json:"-"`
	SentenceID       int64  `json:"-"`
	ID               string `json:"id"`
	Database         string `json:"database,omit_empty"`
	RelatedSentences string `json:"related_sentences,omit_empty"`
	Sentence         string `json:"sentence,omit_empty"`
	SentenceCU       string `json:"sentence_cu,omit_empty"`
	SentenceKeywords string `json:"sentence_keywords,omit_empty"`
	SentenceOriginal string `json:"sentence_original,omit_empty"`
	SentenceSyn      string `json:"sentence_syn,omit_empty"`
	SentenceType     string `json:"sentence_type,omit_empty"`
	SentencePos      string `json:"sentence_pos,omit_empty"`
	Source           string `json:"source,omit_empty"`
	Type             string `json:"type,omit_empty"`
	Enabled          bool   `json:"autofill_enabled,omit_empty"`
}

type RelatedSentence struct {
	Answer  string `json:"answer"`
	CU      string `json:"cu"`
	Creator string `json:"creator,omitempty"`
}

func NewRelatedSentence(sentence string) *RelatedSentence {
	return &RelatedSentence{
		Answer: sentence,
		CU:     "{}",
	}
}

type Sentence struct {
	ModuleID   int64
	SentenceID int64
	Sentence   string
}

type AutofillToggleBody struct {
	ID              string                 `json:"id"`
	AutofillEnabled map[string]interface{} `json:"autofill_enabled"`
}


type BfAccessToken struct {
	UserId             string     `json:"USER_ID"`
	AccessToken        string     `json:"access_token"`
	Expiration         float64      `json:"expiration"`
	CreateDatetime     time.Time  `json:"create_datetime"`
}

func NewAutofillToggleBody(id string, enabled bool) *AutofillToggleBody {
	return &AutofillToggleBody{
		ID: id,
		AutofillEnabled: map[string]interface{}{
			"set": enabled,
		},
	}
}
