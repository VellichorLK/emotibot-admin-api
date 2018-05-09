package data

import (
	"emotibot.com/emotigo/module/admin-api/SelfLearning/internal/model"
)

type DalItem struct {
	Content   string
	ContentID uint64
	Tokens    []string
	KeyWords  []string
	Word2Vec  map[string]model.Vector
	Embedding model.Vector
	Annotated bool
}

type NativeLog struct {
	Logs []*DalItem
}

func (l *NativeLog) Init() {
	l.Logs = make([]*DalItem, 0)
}
