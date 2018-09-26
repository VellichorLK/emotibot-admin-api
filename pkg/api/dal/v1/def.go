package dal

import (
	"strings"
)

type request struct {
	Op           string `json:"op"`
	Category     string `json:"category"`
	AppID        string `json:"appid"`
	UserRecordID int64  `json:"userRecordId"`
	Data         data   `json:"data"`
}

type data struct {
	Subop      string      `json:"subop"`
	Conditions *conditions `json:"conditions,omitempty"`
	Entities   []entity    `json:"entities,omitempty"`
	Groupby    string      `json:"groupby,omitempty"`
	Algo       string      `json:"algo,omitempty"`
	StartIndex int64       `json:"startIndex,omitempty"`
	EndIndex   int64       `json:"endIndex,omitempty"`
	Segments   []string    `json:"segments,omitempty"`
}

type entity struct {
	Content       string `json:"content"`
	ParentContent string `json:"parentContent"`
}

type conditions struct {
	Labels        []string `json:"labels,omitempty"`
	Content       string   `json:"content,omitempty"`
	ParentContent string   `json:"parentContent,omitempty"`
}

type response struct {
	ErrNo      string        `json:"errno"`
	ErrMessage string        `json:"errmsg"`
	Results    []interface{} `json:"actualResults"`
	Operation  []string      `json:"results"`
}

//DetailError should contain detail dal response message for debug.
type DetailError struct {
	ErrMsg  string
	Results []string
}

func (e *DetailError) Error() string {
	msg := "dal error: " + e.ErrMsg
	if e.Results != nil && len(e.Results) > 0 {
		msg += " with result [" + strings.Join(e.Results, ", ") + "]"
	}
	return msg
}
