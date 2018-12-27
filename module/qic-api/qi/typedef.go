package qi

import (
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
