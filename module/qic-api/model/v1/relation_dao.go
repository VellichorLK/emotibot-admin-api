package model

import (
	"database/sql"
	"errors"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

//RelationDao access the relational table
type RelationDao interface {
	GetLevelRelationID(sql SqlLike, from int, to int, id []uint64) ([]map[uint64][]uint64, [][]uint64, error)
}

type relSelectFld struct {
	tblName  string
	fields   [2]string
	nickName string
}

var leveltblMap = map[int]relSelectFld{
	0: relSelectFld{tblName: tblRelGrpRule, fields: [2]string{RGCGroupID, RRRRuleID}, nickName: "a"},
	1: relSelectFld{tblName: tblRelCRCF, fields: [2]string{RRRRuleID, RCFSGCFID}, nickName: "b"},
	2: relSelectFld{tblName: tblRelCFSG, fields: [2]string{RCFSGCFID, RSGSSGID}, nickName: "c"},
	3: relSelectFld{tblName: tblRelSGS, fields: [2]string{RSGSSGID, fldRelSenID}, nickName: "d"},
	4: relSelectFld{tblName: tblRelSenTag, fields: [2]string{fldRelSenID, fldRelTagID}, nickName: "e"},
}

var numOfLevel = len(leveltblMap)

//RelationSQLDao uses the mysql as db
type RelationSQLDao struct {
}

//error message
var (
	ErrOutOfLevel = errors.New("Out of levels")
	ErroNoConn    = errors.New("No connection")
)

//GetLevelRelationID gives the relation ID
//the value range of arguments, from and to, is 0~5
//the arguments of id means the parent id condition
//return value is slice of map which means in each relation table, the parent id contains childs's
//[][]uint64 means the order of each parent id in each relation table
func (d *RelationSQLDao) GetLevelRelationID(delegatee SqlLike, from int, to int, id []uint64) ([]map[uint64][]uint64, [][]uint64, error) {

	if delegatee == nil {
		return nil, nil, ErroNoConn
	}
	if to <= from {
		return nil, nil, nil
	}

	if to > numOfLevel || from < 0 {
		return nil, nil, ErrOutOfLevel
	}

	numOfIDs := len(id)
	if numOfIDs == 0 {
		return nil, nil, nil
	}

	use := to - from
	selectStr := ""
	joinsStr := ""
	orderStr := ""
	var lastLevel relSelectFld

	//compose the selected fields and join tables
	for i := from; i < (use + from); i++ {
		level := leveltblMap[i]
		if i == from {
			joinsStr = level.tblName + " AS " + level.nickName
			selectStr = level.nickName + "." + level.fields[0] + "," + level.nickName + "." + level.fields[1]
			orderStr = level.nickName + "." + fldID + " ASC"
		} else {
			joinsStr = joinsStr + " LEFT JOIN " + level.tblName + " AS " + level.nickName + " ON " +
				lastLevel.nickName + "." + lastLevel.fields[1] + "=" + level.nickName + "." + level.fields[0]
			selectStr = selectStr + "," + level.nickName + "." + level.fields[1]
			orderStr = orderStr + "," + level.nickName + "." + fldID + " ASC"
		}
		lastLevel = level
	}
	fromLevelTbl := leveltblMap[from]
	querySQL := "SELECT " + selectStr + " FROM " + joinsStr +
		" WHERE " + fromLevelTbl.nickName + "." + fromLevelTbl.fields[0] + " IN (?" + strings.Repeat(",?", numOfIDs-1) + ")" +
		" ORDER BY " + orderStr

	//fmt.Printf("%s\n%+v\n", querySQL, id)

	//transform id to interface type
	idInterface := make([]interface{}, 0, numOfIDs)
	for _, v := range id {
		idInterface = append(idInterface, v)
	}

	rows, err := delegatee.Query(querySQL, idInterface...)
	if err != nil {
		logger.Error.Printf("Query error. %s\n%s\n", querySQL, err)
		return nil, nil, err
	}
	defer rows.Close()

	numOfScan1 := use + 1
	numOfRelation := use

	resp := make([]map[uint64][]uint64, numOfRelation, numOfRelation)
	order := make([][]uint64, numOfRelation)
	//records the duplicate id
	recordDup := make([]map[uint64]map[uint64]bool, numOfRelation, numOfRelation)
	relIDs := make([]interface{}, 0, numOfScan1)
	for i := 0; i < numOfScan1; i++ {
		relIDs = append(relIDs, new(sql.NullInt64))
	}

	for rows.Next() {
		err = rows.Scan(relIDs...)
		if err != nil {
			return nil, nil, err
		}

		var lastID uint64
		for k, v := range relIDs {
			val, ok := v.(*sql.NullInt64)
			if !ok {
				lastID = 0
				logger.Error.Printf("transform to *uint64 failed\n")
				continue
			}
			if !val.Valid {
				lastID = 0
				continue
			}
			varUint64 := uint64(val.Int64)
			if lastID == 0 {
				lastID = varUint64
				continue
			}
			ithTbl := k - 1

			if resp[ithTbl] == nil {
				resp[ithTbl] = make(map[uint64][]uint64)
				recordDup[ithTbl] = make(map[uint64]map[uint64]bool)
			}

			if recordDup[ithTbl][lastID] == nil {
				recordDup[ithTbl][lastID] = make(map[uint64]bool)
				order[ithTbl] = append(order[ithTbl], lastID)
			}

			//already in the list
			if recordDup[ithTbl][lastID][varUint64] {
				lastID = varUint64
				continue
			} else {
				recordDup[ithTbl][lastID][varUint64] = true
			}
			resp[ithTbl][lastID] = append(resp[ithTbl][lastID], varUint64)
			lastID = varUint64
		}
	}

	return resp, order, nil
}
