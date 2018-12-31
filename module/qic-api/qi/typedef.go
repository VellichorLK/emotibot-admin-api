package qi

import (
	"database/sql"
	"encoding/json"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

type GroupsResponse struct {
	Paging *general.Paging    `json:"paging"`
	Data   []model.GroupWCond `json:"data"`
}

type SimpleGroupsResponse struct {
	Paging *general.Paging     `json:"paging"`
	Data   []model.SimpleGroup `json:"data"`
}

type SentenceGroupInReq struct {
	ID               string   `json:"sg_id"`
	Name             string   `json:"sg_name"`
	Role             string   `json:"role"`
	Position         string   `json:"position"`
	PositionDistance int      `json:"position_distance"`
	Sentences        []string `json:"sentences"`
}

type SentenceGroupInResponse struct {
	ID               string                 `json:"sg_id,omitempty"`
	Name             string                 `json:"sg_name,omitempty"`
	Role             string                 `json:"role,omitempty"`
	Position         string                 `json:"position,omitempty"`
	PositionDistance int                    `json:"position_distance"`
	Sentences        []model.SimpleSentence `json:"sentences,omitempty"`
}

//TagDao is tag resource manipulating interface, which itself should support ACID transaction.
type TagDao interface {
	Begin() (*sql.Tx, error)
	Tags(tx *sql.Tx, query model.TagQuery) ([]model.Tag, error)
	NewTags(tx *sql.Tx, tags []model.Tag) ([]model.Tag, error)
	DeleteTags(tx *sql.Tx, query model.TagQuery) (int64, error)
	CountTags(tx *sql.Tx, query model.TagQuery) (uint, error)
}

//Query parameter
const (
	QPage  = "page"
	QLimit = "limit"
)

//Default value
const (
	DPage  = 1
	DLimit = 10
)

//TagResponse is the Get handler response body struct.
type TagResponse struct {
	Paging general.Paging `json:"paging"`
	Data   []tag          `json:"data"`
}

type tag struct {
	TagID        uint64          `json:"tag_id"`
	TagName      string          `json:"tag_name"`
	TagType      string          `json:"tag_type"`
	PosSentences json.RawMessage `json:"pos_sentences"`
	NegSentences json.RawMessage `json:"neg_sentences"`
}
