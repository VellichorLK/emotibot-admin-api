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
