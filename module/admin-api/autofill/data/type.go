package data

import (
	qaData "emotibot.com/emotigo/module/admin-api/QADoc/data"
)

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
}

type Sentence struct {
	ModuleID   int64
	SentenceID int64
	Sentence   string
}
