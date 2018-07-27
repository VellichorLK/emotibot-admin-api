package v2

// SentenceV2 describe a sentence in trainint data
type SentenceV2 struct {
	ID      int64  `json:"id"`
	Content string `json:"count"`
}

// SentenceV2WithType describe a sentence in trainint data
type SentenceV2WithType struct {
	SentenceV2
	Type int `json:"type"`
}

// IntentV2 describe a intent in V2
type IntentV2 struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	PositiveCount int            `json:"positive_count"`
	NegativeCount int            `json:"negative_count"`
	Positive      *[]*SentenceV2 `json:"positive,omitempty"`
	Negative      *[]*SentenceV2 `json:"negative,omitempty"`
}
