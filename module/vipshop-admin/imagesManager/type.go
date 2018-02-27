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

//table name of meida base center
const (
	imageTable    = "images"
	locationTable = "image_location"
	relationTable = "image_question"
)

//field name of each table
const (
	attrID           = "id"
	attrFileName     = "fileName"
	attrLocationID   = "locationId"
	attrCreateTime   = "createdTime"
	attrLatestUpdate = "lastModified"

	attrLocation   = "location"
	attrImageID    = "image_id"
	attrQuestionID = "question_id"
)

//request parameter name
const ()

//error number of mysql
const (
	ErDupEntry = 1062
)
