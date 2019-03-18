package qi

import (
	"time"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	callGroupDao model.CallGroupDao = &model.CallGroupSQLDao{}
)

// CreateCallGroupCondition create a new call group condition
func CreateCallGroupCondition(reqModel *model.CallGroupCondition, enterprise string) (int64, error) {
	if reqModel == nil {
		return 0, ErrNoArgument
	}
	if dbLike == nil {
		return 0, ErrNilCon
	}
	reqModel.Enterprise = enterprise
	reqModel.CreateTime = time.Now().Unix()
	reqModel.UpdateTime = reqModel.CreateTime
	id, err := callGroupDao.CreateCondition(dbLike.Conn(), reqModel)
	return id, err
}

// GetCallGroupConditionList return the detail of a call group condition
func GetCallGroupConditionList(query *model.GeneralQuery, pagination *model.Pagination) ([]*model.CallGroupCondition, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	return callGroupDao.GetConditionList(dbLike.Conn(), query, pagination)
}

//CountCallGroupCondition counts the total number of call group condition
func CountCallGroupCondition(query *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	return callGroupDao.CountCondition(dbLike.Conn(), query)
}

// UpdateCallGroupCondition update the call group condition
func UpdateCallGroupCondition(query *model.GeneralQuery, data *model.CallGroupConditionUpdateSet) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNoID
	}
	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("create session failed. %s\n", err)
		return 0, err
	}
	defer tx.Rollback()

	id, err := callGroupDao.UpdateCondition(tx, query, data)
	if err != nil {
		logger.Error.Printf("update failed. %s\n", err)
		return 0, err
	}
	tx.Commit()
	return id, nil

}

// DeleteCallGroupCondition delete the call group condition
func DeleteCallGroupCondition(query *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNoID
	}
	return callGroupDao.SoftDeleteCondition(dbLike.Conn(), query)
}
