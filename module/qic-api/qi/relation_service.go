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
	ErrWrongLevel = errors.New("Wrong level assigned")
	ErrOutOfLevel = errors.New("assigned level out of level")
	ErrNoID       = errors.New("Must has id")
)

// GetLevelsRel assemble a relationTable from the from level to to level.
// from and to indicate which level resource(ex: from LevRuleGroup, to LevTag. then it will check )
// from must be lower than to, or a ErrWrongLevel will returned.
// id is the from resource's primary key id. which will be the first level of map key
// return values:
//	- relationTable is a slice of map in each relation table. key is the from level's primary key id,
//	  value is the to level's primary key id.
//		ex:
//		input: from RuleGroup to conversation, id [1]
//		[]{
//			RuleGroup -> rule
// 			map{
//				1: [1, 2],
//			}
//			rule -> conversation
//			map{
//				1: [1, 3, 5],
//				2: [2, 4],
//			}
//		}
//  - order is the the order of each parent id in each relation table.
//  - err will be nil if success. if to or from level is a invalid number, a ErrOutOfLevel is returned.
//	  If id is empty, ErrNoID is returned.
func GetLevelsRel(from Levels, to Levels, id []uint64) (relationTable []map[uint64][]uint64, order [][]uint64, err error) {
	if to <= from {
		return nil, nil, ErrWrongLevel
	}
	if to > LevTag || from < LevRuleGroup {
		return nil, nil, ErrOutOfLevel
	}
	if len(id) == 0 {
		return nil, nil, ErrNoID
	}
	return relationDao.GetLevelRelationID(dbLike.Conn(), int(from), int(to), id)
}
