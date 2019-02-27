package service

import (
	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"errors"
	_ "database/sql"
)

var useDBHX dao.DBHX

// SetDBV4 will setup a db implement interface of dao.DBV4
func SetDBHX(db dao.DBHX) {
	useDBHX = db
}

func checkDBHX() error {
	if useDBHX == nil {
		return errors.New("DB hasn't set")
	}
	return nil
}
func GetAllRoles(enterpriseID string)([]*data.RoleHX, error) {
	err := checkDBHX()
	if err != nil {
		return nil, err
	}
	return useDBHX.GetRolesHX(enterpriseID)

}

func GetAllModules()(map[string]interface{}, error) {
	err := checkDBHX()
	if err != nil {
		return nil, err
	}
	return useDBHX.GetModulesHX()

}

func GetRolePrivileges(enterpriseID string,roleId int)(map[string]interface{}, error) {
	err := checkDBHX()
	if err != nil {
		return nil, err
	}
	return useDBHX.GetRolePrivileges(enterpriseID,roleId)

}

func UpdateRolePrivileges(enterpriseID string,roleId int,privileges map[string]map[string]map[string][]string)(error) {
	err := checkDBHX()
	if err != nil {
		return err
	}
	return useDBHX.UpdateRolePrivileges(enterpriseID,roleId,privileges)

}