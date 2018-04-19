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
	ID               int       `json:"id"`
	Question         string    `json:"main_question"`
	RelatedQuestions []string  `json:"relate_questions"`
	Answers          []string  `json:"answers"`
	CreatedTime      time.Time `json:"created_time"`
}

// RetQAInfo is the struct in api return
type RetQAInfo struct {
	Count int       `json:"count"`
	Infos []*QAInfo `json:"qa_infos"`
}

// ChatInfo store info about robot chat setting
type ChatInfo struct {
	Type     int      `json:"type"`
	Contents []string `json:"contents"`
}

type ChatDescription struct {
	Type    int    `json:"type"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

// ChatInfoInput is used when update robot chat setting
type ChatInfoInput struct {
	Type     int      `json:"type"`
	Contents []string `json:"contents"`
	Name     string   `json:"name"`
}

type Function struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Remark string `json:"remark"`
	Intent string `json:"intent"`
}
