package v1

import (
	"strings"
)

type request struct {
	Op               string         `json:"op"`
	Category         string         `json:"category"`
	AppID            string         `json:"appId"`
	SqContent        string         `json:"sqContent"`
	LqContent        string         `json:"lqContent"`
	LqContentList    []string       `json:"lqContentList"`
}

type SQ struct {
	AppId          string         `json:"app_id"`
	Content        string         `json:"content"`
	CreateTime     string         `json:"create_time"`
	CreateUser     string         `json:"create_user"`
	Id             int64          `json:"id"`
	UpdateTime     string         `json:"update_time"`
	UpdateUser     string         `json:"update_user"`
}

type LQ struct {
	LqContent      string         `json:"lqcontent"`
	Lqid           int64          `json:"lqid"`
	SqContent      string         `json:"sqcontent"`
	Sqid           int64          `json:"sqid"`
}

type SQResult struct {
	ActualResults  []SQ    `json:"actualResults"`
	Errno          string  `json:"errno"`
}

type LQResult struct {
	ActualResults  []LQ    `json:"actualResults"`
	Errno          string  `json:"errno"`
}


// RawResponse is the general response retrive from dal api endpoint.
// It is exported only for test package to avoid duplicated struct.
type RawResponse struct {
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

