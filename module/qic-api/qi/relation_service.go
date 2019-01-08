package qi

import (
	"errors"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

//Levels defines the level of the whole qi check conversation
type Levels int

//each level of the whole picture
const (
	LevRuleGroup Levels = iota
	LevRule
	LevConversation
	LevSenGroup
	LevSentence
	LevTag
)

var (
	relationDao model.RelationDao
)

//error message
var (
	ErrWrongLevel = errors.New("Wrong error assigned")
	ErrOutOfLevel = errors.New("assigned level out of level")
	ErrNoID       = errors.New("Must has id")
)

//GetLevelsRel gives  from the from level to to level.
//id is the parent_id which is the second field in from level table
//return value is slice of map which means in each relation table, the parent id contains childs's
func GetLevelsRel(from Levels, to Levels, id []uint64) ([]map[uint64][]uint64, error) {
	if to <= from {
		return nil, ErrWrongLevel
	}
	if to > LevTag || from < LevRuleGroup {
		return nil, ErrOutOfLevel
	}
	if len(id) == 0 {
		return nil, ErrNoID
	}
	return relationDao.GetLevelRelationID(dbLike.Conn(), int(from), int(to), id)
}
