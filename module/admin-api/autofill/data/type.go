package data

import (
	qaData "emotibot.com/emotigo/module/admin-api/QADoc/data"
)
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

type QACoreDoc struct {
	qaData.QACoreDoc
	ModuleID   int64
	SentenceID int64
	Sentence   string
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
	Expiration         float64    `json:"expiration"`
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
