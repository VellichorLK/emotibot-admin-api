package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

// CreditCallGroupDao defines the dao interface of grouped prediction result (CUPredictResultGroup)
type CreditCallGroupDao interface {
	CreateCreditCallGroup(conn SqlLike, model *CreditCallGroup) (int64, error)
	UpdateCreditCallGroup(conn SqlLike, query *GeneralQuery, data *CreditCallGroupUpdateSet) (int64, error)
}

// CreditCallGroupSQLDao defines SQL implementation of CreditCallGroupDao
type CreditCallGroupSQLDao struct {
}

//CreditCallGroup defines the model struture of CUPredictResultGroup
type CreditCallGroup struct {
	ID          uint64
	CallGroupID uint64
	Type        int
	ParentID    uint64
	OrgID       uint64
	Valid       int
	Revise      int
	Score       int
	CreateTime  int64
	UpdateTime  int64
	CallID      uint64
}

//CreateCreditCallGroup careate a new CreditCallGroup record
func (*CreditCallGroupSQLDao) CreateCreditCallGroup(conn SqlLike, model *CreditCallGroup) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if model == nil {
		return 0, ErrNeedRequest
	}
	insertCols := []string{
		fldCallGroupID, fldType, fldParentID, fldOrgID, fldValid,
		fldRevise, fldScore, fldCreateTime, fldUpdateTime, fldCallID,
	}
	params := []interface{}{
		model.CallGroupID, model.Type, model.ParentID, model.OrgID, model.Valid,
		model.Revise, model.Score, model.CreateTime, model.UpdateTime, model.CallID,
	}

	querySQL := fmt.Sprintf(
		"INSERT INTO `%s` (`%s`) VALUES (%s)",
		tblPredictResultGroup,
		strings.Join(insertCols, "`, `"),
		"?"+strings.Repeat(",?", len(insertCols)-1),
	)

	result, err := conn.Exec(querySQL, params...)
	if err != nil {
		logger.Error.Printf("sql execution failed. %s %+v\n", querySQL, params)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, ErrAutoIDDisabled
	}
	return id, nil
}

// CreditCallGroupUpdateSet defines the json body of handleUpdateCallGroupCondition request
type CreditCallGroupUpdateSet struct {
	ParentID   *uint64
	OrgID      *uint64
	Valid      *int
	Revise     *int
	Score      *int
	UpdateTime *int64
	CallID     *uint64
}

//UpdateCreditCallGroup updates the content of CreditCallGroup
func (*CreditCallGroupSQLDao) UpdateCreditCallGroup(conn SqlLike, query *GeneralQuery, data *CreditCallGroupUpdateSet) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldParentID, fldOrgID, fldValid, fldRevise, fldScore,
		fldUpdateTime, fldCallID,
	}
	return updateSQL(conn, query, data, tblPredictResultGroup, flds)
}
