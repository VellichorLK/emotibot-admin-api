package qi

import (
	"database/sql"

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
}

type SentenceGroupInResponse struct {
	ID               string                 `json:"sg_id"`
	Name             string                 `json:"sg_name"`
	Role             string                 `json:"role"`
	Position         string                 `json:"position"`
	PositionDistance int                    `json:"position_distance"`
	Sentences        []model.SimpleSentence `json:"sentences"`
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
	TagUUID      string   `json:"tag_id,omitempty"`
	TagName      string   `json:"tag_name,omitempty"`
	TagType      string   `json:"tag_type,omitempty"`
	PosSentences []string `json:"pos_sentences,omitempty"`
	NegSentences []string `json:"neg_sentences,omitempty"`
}

type CallDao interface {
	Calls(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error)
	NewCalls(delegatee model.SqlLike, calls []model.Call) ([]model.Call, error)
	SetRuleGroupRelations(delegatee model.SqlLike, call model.Call, rulegroups []model.Group) ([]int64, error)
	SetCall(delegatee model.SqlLike, call model.Call) error
	Count(delegatee model.SqlLike, query model.CallQuery) (int64, error)
}

type TaskDao interface {
	CallTask(delegatee model.SqlLike, call model.Call) (model.Task, error)
	NewTask(delegatee model.SqlLike, task model.Task) (*model.Task, error)
}

type SegmentDao interface {
	NewSegments(delegatee model.SqlLike, segments []model.RealSegment) ([]model.RealSegment, error)
	Segments(delegatee model.SqlLike, query model.SegmentQuery) ([]model.RealSegment, error)
}

type controllerError struct {
	error
	errNo int
}

type adminError interface {
	error
	ErrorNo() int
}
