package imagesManager

import "emotibot.com/emotigo/module/vipshop-admin/util"

//Image represent sql schema of the table image
//Note: It does not contain bytes stream of image, instead it should be read from the location
type Image struct {
	ID               int64
	FileName         string
	Location         string
	CreatedTime      util.JSONUnixTime
	LastModifiedTime util.JSONUnixTime
}

type uploadArg struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

type getImagesArg struct {
	Order   string
	Page    int64
	Limit   int64
	Keyword string
}

type imageList struct {
	Total     uint64       `json:"total"`
	CurPage   uint64       `json:"curPage"`
	Images    []*imageInfo `json:"Images"`
	answerIDs []interface{}
}

type imageInfo struct {
	ImageID      uint64          `json:"id"`
	Title        string          `json:"title"`
	Size         int             `json:"size"`
	CreateTime   uint64          `json:"createTime"`
	LastModified uint64          `json:"lastModified"`
	Refs         []*questionInfo `json:"refs"`
	URL          string          `json:"url"`
	answerID     []int
}

type questionInfo struct {
	QuestionID int    `json:"questionId"`
	Info       string `json:"info"`
}

//table name of meida base center
const (
	imageTable    = "images"
	locationTable = "image_location"
	relationTable = "image_answer"
)

//field name of each table
const (
	attrID           = "id"
	attrFileName     = "fileName"
	attrSize         = "size"
	attrLocationID   = "locationId"
	attrCreateTime   = "createdTime"
	attrLatestUpdate = "lastModified"

	attrLocation   = "location"
	attrImageID    = "image_id"
	attrQuestionID = "question_id"
	attrAnswerID   = "answer_id"
)

//vipshop table attribute
const (
	attrQID          = "Question_Id"
	attrTag          = "Tags"
	attrTagID        = "Tag_Id"
	attrTagName      = "Tag_Name"
	attrAnsID        = "Answer_Id"
	attrContent      = "Content"
	attrCategoryID   = "CategoryId"
	attrCategoryName = "CategoryName"
	attrParentID     = "ParentId"
)

//vipsho table name
const (
	VIPAnswerTable    = "vipshop_answer"
	VIPAnswerTagTable = "vipshop_answertag"
	VIPCategoryTable  = "vipshop_categories"
	VIPQuestionTable  = "vipshop_question"
	VIPTagTable       = "vipshop_tag"
)

//request parameter name
const (
	ORDER   = "order"
	PAGE    = "page"
	LIMIT   = "limit"
	KEYWORD = "keyword"
)

//request parameter value
const (
	valID   = "id"
	valName = "name"
	valTime = "time"
)

//error number of mysql
const (
	ErDupEntry = 1062
)

//Category record category name and its parent id
type Category struct {
	Name     string
	ParentID int
}
