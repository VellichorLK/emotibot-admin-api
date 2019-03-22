package service

import (
	"database/sql"
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

func GetUserPrivileges(enterpriseID string,userCode string)(map[string]map[string]map[string][]string, error) {
	err := checkDBHX()
	if err != nil {
		return nil, err
	}
	return useDBHX.GetUserPrivileges(enterpriseID,userCode)

}

func GetRolePrivileges(enterpriseID string,roleId int)(map[string]map[string]map[string][]string, error) {
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
func GetLabelUsers(enterpriseID string,role int)([]*data.UserHX,error) {
	err := checkDBHX()
	if err != nil {
		return nil,err
	}
	return useDBHX.GetLabelUsers(enterpriseID,role)

}

func GetUserAccessInfo(userCode string) (*data.UserHX, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}
	info, err := useDBHX.GetUserAccessInfo(userCode)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return info, nil
}