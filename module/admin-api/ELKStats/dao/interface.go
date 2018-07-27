package dao

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
)

// DB defines interface for different DAO modules
type DB interface {
	GetTags() (map[string][]data.Tag, error)
}
