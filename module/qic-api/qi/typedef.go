package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

type GroupsResponse struct {
	Paging *general.Paging    `json:"paging"`
	Data   []model.GroupWCond `json:"data"`
}
type CallsResponse struct {
	Paging general.Paging `json:"paging"`
	Data   []CallResp     `json:"data"`
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
	Type             string   `json:"type"`
	Optional         bool     `json:"optional"`
}

type SentenceGroupInResponse struct {
	ID               string                 `json:"sg_id"`
	Name             string                 `json:"sg_name"`
	Role             string                 `json:"role"`
	Position         string                 `json:"position"`
	PositionDistance int                    `json:"position_distance"`
	Sentences        []model.SimpleSentence `json:"sentences"`
}

//Query parameter
const (
	QPage  = "page"
	QLimit = "limit"
)

//Default value
const (
	DPage = 1
)

//TagResponse is the Get handler response body struct.
type TagResponse struct {
	Paging general.Paging `json:"paging"`
	Data   []tag          `json:"data"`
}

type tag struct {
	TagUUID      string   `json:"tag_id,omitempty"`
	TagName      string   `json:"tag_name,omitempty"`
	TagType      string   `json:"tag_type,omitempty"`
	PosSentences []string `json:"pos_sentences,omitempty"`
	NegSentences []string `json:"neg_sentences,omitempty"`
}

type controllerError struct {
	error
	errNo int
}

type adminError interface {
	error
	ErrorNo() int
}
