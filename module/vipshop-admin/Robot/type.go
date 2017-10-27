package Robot

import (
	"time"
)

// FunctionInfo store info about robot's function
type FunctionInfo struct {
	Status bool `json:"status"`
}

// QAInfo store info about robot's qa pair
// First question in questions is main question
type QAInfo struct {
	ID          int       `json:"id"`
	Questions   []string  `json:"questions"`
	Answers     []string  `json:"answers"`
	CreatedTime time.Time `json:"created_time"`
}

// RetQAInfo is the struct in api return
type RetQAInfo struct {
	Count int       `json:"count"`
	Infos []*QAInfo `json:"qa_infos"`
}
