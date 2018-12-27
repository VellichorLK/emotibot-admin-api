package qi

import (
	"database/sql"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util"
)

type GroupsResponse struct {
	Paging *util.Paging       `json:"paging"`
	Data   []model.GroupWCond `json:"data"`
}

type SimpleGroupsResponse struct {
	Paging *util.Paging        `json:"paging"`
	Data   []model.SimpleGroup `json:"data"`
}

//TagDao is tag resource manipulating interface, which itself should support ACID transaction.
type TagDao interface {
	Tags(tx *sql.Tx, query model.TagQuery) ([]model.Tag, error)
	NewTags(tx *sql.Tx, tags []model.Tag) ([]uint, error)
	SetTags(tx *sql.Tx, tag []model.Tag) ([]uint, error)
	DeleteTags(tx *sql.Tx, query model.TagQuery) error
}
