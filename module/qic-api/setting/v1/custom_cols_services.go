package setting

import (
	"fmt"
	"time"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

// a declaration to outside dependencies
var (
	newUserKey    = userkeyDao.NewUserKey
	userKeys      = userkeyDao.UserKeys
	countUserKeys = userkeyDao.CountUserKeys
	deleteUserKey = userkeyDao.DeleteUserKeys
)

// NewUKRequest is the request for create a user key record.
type NewUKRequest struct {
	Enterprise string
	Name       string
	InputName  string
	Type       int8
}

var defaultCustomColService = struct {
	GetCustomCols    func(query model.UserKeyQuery) ([]CustomCol, general.Paging, error)
	NewCustomCols    func(requests []NewUKRequest) ([]model.UserKey, error)
	DeleteCustomCols func(enterprise string, inputnames ...string) (int64, error)
}{
	GetCustomCols: func(query model.UserKeyQuery) ([]CustomCol, general.Paging, error) {
		userKeys, err := userKeys(nil, query)
		if err != nil {
			return nil, general.Paging{}, fmt.Errorf("get user keys failed, %v", err)
		}
		cols := []CustomCol{}
		for _, key := range userKeys {
			cols = append(cols, CustomCol{
				ID:        key.ID,
				Name:      key.Name,
				InputName: key.InputName,
				Type:      key.Type,
			})
		}
		size, err := countUserKeys(nil, query)
		page := general.Paging{
			Page:  query.Paging.Page,
			Limit: query.Paging.Limit,
			Total: size,
		}
		return cols, page, nil
	},
	NewCustomCols: func(requests []NewUKRequest) ([]model.UserKey, error) {
		tx, err := db.Begin()
		if err != nil {
			return nil, fmt.Errorf("db begin failed, %v", err)
		}
		defer tx.Rollback()
		now := time.Now().Unix()
		createdKeys := []model.UserKey{}
		for _, req := range requests {
			keys, err := userKeys(tx, model.UserKeyQuery{
				Enterprise: req.Enterprise,
				InputNames: []string{req.InputName},
			})
			if err != nil {
				return nil, fmt.Errorf("query user key failed, %v", err)
			}
			if len(keys) > 0 {
				return nil, fmt.Errorf("User Key '%s' already exist", req.InputName)
			}
			key, err := newUserKey(tx, model.UserKey{
				Enterprise: req.Enterprise,
				Name:       req.Name,
				InputName:  req.InputName,
				Type:       req.Type,
				IsDeleted:  false,
				CreateTime: now,
				UpdateTime: now,
			})
			if err != nil {
				return nil, fmt.Errorf("create UserKey failed, %v", err)
			}
			createdKeys = append(createdKeys, key)
		}
		tx.Commit()
		return createdKeys, nil
	},
	DeleteCustomCols: func(enterprise string, inputnames ...string) (int64, error) {
		if enterprise == "" {
			return 0, fmt.Errorf("enterprise is empty")
		}
		if len(inputnames) == 0 {
			return 0, fmt.Errorf("require at least one inputnames")
		}
		query := model.UserKeyQuery{
			InputNames: inputnames,
			Enterprise: enterprise,
		}
		total, err := deleteUserKey(nil, query)
		if err != nil {
			return 0, fmt.Errorf("delete user key failed, %v", err)
		}
		return total, nil
	},
}

// GetCustomCols get user keys from dao, and transform its as the response CustomCol.
func GetCustomCols(query model.UserKeyQuery) ([]CustomCol, general.Paging, error) {
	return defaultCustomColService.GetCustomCols(query)
}

// NewCustomCols create new user key by the request.
func NewCustomCols(requests []NewUKRequest) ([]model.UserKey, error) {
	return defaultCustomColService.NewCustomCols(requests)
}

// DeleteCustomCols delete
func DeleteCustomCols(enterprise string, inputnames ...string) (int64, error) {
	return defaultCustomColService.DeleteCustomCols(enterprise, inputnames...)
}
