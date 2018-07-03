package Service

import (
	"fmt"
	"strings"
)

type NLUSegment struct {
	Word string `json:"word"`
	Pos  string `json:"pos"`
}
type NLUSegments []*NLUSegment

func (segment *NLUSegment) ToString() string {
	return fmt.Sprintf("%s/%s", segment.Word, segment.Pos)
}
func (segments NLUSegments) ToFullString() string {
	temp := make([]string, 0, len(segments))
	for _, seg := range segments {
		temp = append(temp, seg.ToString())
	}
	return strings.Join(temp, " ")
}
func (segments NLUSegments) ToString() string {
	temp := make([]string, 0, len(segments))
	for _, seg := range segments {
		temp = append(temp, seg.Word)
	}
	return strings.Join(temp, " ")
}

type NLUKeyword struct {
	Word  string `json:"word"`
	Pos   string `json:"pos"`
	Level string `json:"level"`
}
type NLUKeywords []*NLUKeyword

func (segment *NLUKeyword) ToString() string {
	return fmt.Sprintf("%s/%s", segment.Word, segment.Pos)
}
func (segments NLUKeywords) ToFullString() string {
	temp := make([]string, 0, len(segments))
	for _, seg := range segments {
		temp = append(temp, seg.ToString())
	}
	return strings.Join(temp, " ")
}
func (segments NLUKeywords) ToString() string {
	temp := make([]string, 0, len(segments))
	for _, seg := range segments {
		temp = append(temp, seg.Word)
	}
	return strings.Join(temp, " ")
}

type NLUResult struct {
	Sentence     string      `json:"query"`
	NLPResult    int         `json:"nlpState"`
	StateCode    int         `json:"stateCode"`
	SentenceType string      `json:"sentenceType"`
	Segment      NLUSegments `json:"segment"`
	Keyword      NLUKeywords `json:"keyword"`
	NLUVersion   string      `json:"jarVersion"`
}
