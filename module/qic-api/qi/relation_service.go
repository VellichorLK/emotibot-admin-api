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

//LevelVaild  is struct of vaild check
type LevelVaild struct {
	Valid       bool
	InValidInfo []*InvalidUnit //if valid is false, then it contains the invalid information
}

//InvalidUnit records the invalid information
type InvalidUnit struct {
	InValidLevel Levels
	InValidID    uint64
}

//error message
var (
	ErrWrongLevel   = errors.New("Wrong level assigned")
	ErrOutOfLevel   = errors.New("assigned level out of level")
	ErrNoID         = errors.New("Must has id")
	ErrUnsupported  = errors.New("Unsupported level")
	ErrLessRelation = errors.New("get less relation level")
)

// GetLevelsRel assemble a relationTable from the from level to to level.
// from and to indicate which level resource(ex: from LevRuleGroup, to LevTag. then it will check )
// from must be lower than to, or a ErrWrongLevel will returned.
// id is the from resource's primary key id. which will be the first level of map key
// if ignoreNULL is true, only the completness level would return. Otherwise, incomplete id would only contain element with 0 value
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
func GetLevelsRel(from Levels, to Levels, id []uint64, ignoreNULL bool) (relationTable []map[uint64][]uint64, order [][]uint64, err error) {
	if to <= from {
		return nil, nil, ErrWrongLevel
	}
	if to > LevTag || from < LevRuleGroup {
		return nil, nil, ErrOutOfLevel
	}
	if len(id) == 0 {
		return nil, nil, ErrNoID
	}
	return relationDao.GetLevelRelationID(dbLike.Conn(), int(from), int(to), id, ignoreNULL)
}

//CheckIntegrity checks the integrity of given level with its id
//return values:
//[]LevelValid for each id, indicates valid or not, if the Valid in LevelValid is false, then InValidInfo is meaningful
func CheckIntegrity(level Levels, id []uint64) ([]LevelVaild, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	if level == LevTag {
		return nil, ErrUnsupported
	}
	numOfIDs := len(id)
	if numOfIDs == 0 {
		return nil, ErrNoID
	}
	rel, _, err := relationDao.GetLevelRelationID(dbLike.Conn(), int(level), int(LevTag), id, false)
	if err != nil {
		return nil, err
	}

	if len(rel) != int(LevTag-level) {
		return nil, ErrLessRelation
	}

	resp := make([]LevelVaild, numOfIDs, numOfIDs)
	levelVaildMap := make(map[uint64]*LevelVaild)
	for k, v := range id {
		levelVaildMap[v] = &resp[k]
		resp[k].Valid = true
	}
	allChildToParent := make([]map[uint64][]uint64, 0, len(rel))
	for lev, pContainsC := range rel {

		//if it's the root level, the invalid case is no id is found
		if lev == 0 {
			for _, v := range id {
				if _, ok := pContainsC[v]; !ok {
					levelVaildMap[v].Valid = false
					invalidInfo := InvalidUnit{InValidLevel: level, InValidID: v}
					levelVaildMap[v].InValidInfo = append(levelVaildMap[v].InValidInfo, &invalidInfo)
				}
			}
		} else {
			for p, cList := range pContainsC {
				//no child, which means the p is not integrity
				if len(cList) == 1 && cList[0] == 0 {
					invalidInfo := InvalidUnit{InValidLevel: Levels(lev) + level, InValidID: p}
					parentIDs := []uint64{p}

					//find the root parentID
					for i := lev - 1; i >= 0; i-- {
						var invalidParents []uint64
						childToParent := allChildToParent[i]
						for _, parentID := range parentIDs {
							invalidParents = append(invalidParents, childToParent[parentID]...)
						}
						parentIDs = invalidParents
					}
					//make the root parent to be not integrity
					for _, parentID := range parentIDs {
						levelVaildMap[parentID].InValidInfo = append(levelVaildMap[parentID].InValidInfo, &invalidInfo)
						levelVaildMap[parentID].Valid = false
					}
				}
			}
		}

		//construct the map to record the slice of parentID to each childID in this level
		//because child may belong to many different parent
		childToParent := make(map[uint64][]uint64)
		for p, cList := range pContainsC {
			for _, c := range cList {
				childToParent[c] = append(childToParent[c], p)
			}
		}
		allChildToParent = append(allChildToParent, childToParent)
	}

	return resp, nil
}
