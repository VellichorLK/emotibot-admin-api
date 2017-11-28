package CAuth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/kataras/iris"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

const (
	applicationName     = "VCA"
	rolesEntry          = "roleRest/getAllRolesByAppName"
	privOfRoleEntry     = "privilegeRest/getPrivilegesByRole"
	usersOfRoleEntry    = "userRest/getUsesByRole"
	roleOfUserEntry     = "userRoleRest/getRolesByUsers"
	removeUserRoleEntry = "userRoleRest/delUserRole"
	addUserRoleEntry    = "userRoleRest/addUserRole"
	removeRolePrivEntry = "rolePrivilegeRest/delRolePrivilege"
	addRolePrivEntry    = "rolePrivilegeRest/addRolePrivilege"
	createRoleEntry     = "roleRest/createRole"
	deleteRoleEntry     = "roleRest/deleteRole"
)

func getRolesFromCAuth() (*AllRolesRet, error) {
	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), rolesEntry)
	param := RolesParam{
		ApplicationName: applicationName,
		AppKey:          getCAuthAppKey(),
	}
	body, err := util.HTTPPostJSON(postURL, param, 5)
	if err != nil {
		fmt.Printf("Get roles from cauth fail: %s\n", err.Error())
		return nil, err
	}

	ret := &AllRolesRet{}
	err = json.Unmarshal([]byte(body), ret)
	if err != nil {
		fmt.Printf("return: %s, err: %s\n", body, err.Error())
		return nil, err
	}
	return ret, nil
}

func getPrivilegeOfRoleFromCAuth(name string) (*PrivilegesRet, error) {
	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), privOfRoleEntry)
	param := RolePrivilegesParam{
		RoleName:        name,
		ApplicationName: applicationName,
		AppKey:          getCAuthAppKey(),
	}
	body, err := util.HTTPPostJSON(postURL, param, 5)
	if err != nil {
		fmt.Printf("Get roles from cauth fail: %s\n", err.Error())
		return nil, err
	}

	ret := &PrivilegesRet{}
	err = json.Unmarshal([]byte(body), ret)
	if err != nil {
		fmt.Printf("return: %s, err: %s\n", body, err.Error())
		return nil, err
	}
	return ret, nil
}

func getUsersOfRoleFromCAuth(name string) (*UsersRet, error) {
	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), usersOfRoleEntry)
	param := RoleUsersParam{
		RoleName:        name,
		ApplicationName: applicationName,
		AppKey:          getCAuthAppKey(),
	}
	body, err := util.HTTPPostJSON(postURL, param, 5)
	if err != nil {
		fmt.Printf("Get users of role from cauth fail: %s\n", err.Error())
		return nil, err
	}

	ret := &UsersRet{}
	err = json.Unmarshal([]byte(body), ret)
	if err != nil {
		fmt.Printf("return: %s, err: %s\n", body, err.Error())
		return nil, err
	}
	return ret, nil
}

func getCAuthRet(url string, param interface{}, receiver interface{}) error {
	body, err := util.HTTPPostJSON(url, param, 5)
	if err != nil {
		fmt.Printf("Get users of role from cauth fail: %s\n", err.Error())
		return err
	}

	err = json.Unmarshal([]byte(body), receiver)
	if err != nil {
		fmt.Printf("return: %s, err: %s\n", body, err.Error())
		return err
	}
	return nil
}

func getCAuthRetWithStatus(url string, param interface{}) (int, string, error) {
	return util.HTTPPostJSONWithStatus(url, param, 5)
}

func GetUserRoles(userID string) ([]*SimpleRoleRet, error) {
	rets, err := getUsersRoles([]string{userID})
	if err != nil {
		return nil, err
	}

	ret := rets.Data[userID]
	if ret == nil || len(ret) <= 0 {
		util.LogTrace.Printf("Get role list: %#v", ret)
		return []*SimpleRoleRet{}, nil
	}
	return ret, nil
}

func getUsersRoles(userIDs []string) (*UserRolesRet, error) {
	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), roleOfUserEntry)
	log.Println(postURL)
	param := UserRolesParam{
		UserAccounts:    userIDs,
		ApplicationName: applicationName,
		AppKey:          getCAuthAppKey(),
	}

	ret := &UserRolesRet{}
	err := getCAuthRet(postURL, param, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func updateUserRole(requester string, userID string, origRoles []*SimpleRoleRet, newRoleID string) error {
	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), addUserRoleEntry)
	param := UserRoleInput{
		RoleName:        newRoleID,
		UserAccount:     userID,
		Requestor:       requester,
		ApplicationName: applicationName,
		AppKey:          getCAuthAppKey(),
	}

	deleteRoles := []*SimpleRoleRet{}
	newRoleExisted := false

	for _, role := range origRoles {
		if role.RoleName != newRoleID {
			deleteRoles = append(deleteRoles, role)
		} else {
			newRoleExisted = true
		}
	}

	if newRoleID != "" && !newRoleExisted {
		code, body, err := getCAuthRetWithStatus(postURL, param)
		if err != nil {
			return err
		}
		if code != iris.StatusOK {
			apiRet := CAuthStatus{}
			err := json.Unmarshal([]byte(body), &apiRet)
			if err != nil {
				return fmt.Errorf("Add Role fail: %d [%s]", code, body)
			} else {
				return fmt.Errorf("Add Role fail: %s", apiRet.Error.Message)
			}
		}
	}

	postURL = fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), removeUserRoleEntry)
	for _, role := range deleteRoles {
		param.RoleName = role.RoleName
		code, body, err := util.HTTPRequestJSONWithStatus(postURL, param, 5, "DELETE")
		if err != nil {
			return err
		}
		if code != iris.StatusOK {
			return fmt.Errorf("Remove orig role faild: %s", body)
		}

		return nil
	}

	return nil
}

func GetRolePrivs(id string) (map[int][]string, error) {
	cauthPrivSet, err := getPrivilegeOfRoleFromCAuth(id)
	if err != nil {
		return nil, err
	}
	ret := convertCAuthPrivToAPIPriv(cauthPrivSet)
	return ret, nil
}

func updateRolePriv(operator string, roleID string, oldPriv map[int][]string, newPriv map[int][]string) error {
	origCAuthPriv := convertAPIPrivToCAuthPriv(oldPriv)
	newCAuthPriv := convertAPIPrivToCAuthPriv(newPriv)

	deletePrivs := []string{}
	addPrivs := []string{}

	for _, priv := range newCAuthPriv {
		if !util.Contains(origCAuthPriv, priv) {
			addPrivs = append(addPrivs, priv)
		}
	}
	for _, priv := range origCAuthPriv {
		if !util.Contains(newCAuthPriv, priv) {
			deletePrivs = append(deletePrivs, priv)
		}
	}

	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), addRolePrivEntry)
	param := RolePrivilegeInput{
		RoleName:        roleID,
		ApplicationName: applicationName,
		Requestor:       operator,
		AppKey:          getCAuthAppKey(),
	}
	for _, priv := range addPrivs {
		param.PrivilegeName = priv
		ret, body, err := util.HTTPPostJSONWithStatus(postURL, param, 5)
		if err != nil {
			return err
		}
		if ret != iris.StatusOK {
			errMsg := fmt.Sprintf("Add priv %s to role %s fail: %s", priv, roleID, body)
			util.LogInfo.Printf("%s\n", errMsg)
			return errors.New(errMsg)
		}
		util.LogTrace.Printf("Add priv [%s] from [%s]: %s", priv, roleID, body)
	}

	postURL = fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), removeRolePrivEntry)
	for _, priv := range deletePrivs {
		param.PrivilegeName = priv
		ret, body, err := util.HTTPRequestJSONWithStatus(postURL, param, 5, "DELETE")
		if err != nil {
			return err
		}
		if ret != iris.StatusOK {
			errMsg := fmt.Sprintf("delete priv %s to role %s fail: %s", priv, roleID, body)
			util.LogInfo.Printf("%s\n", errMsg)
			return errors.New(errMsg)
		}
		util.LogTrace.Printf("Delete priv [%s] from [%s]: %s", priv, roleID, body)
	}

	return nil
}

func convertCAuthPrivToAPIPriv(CAuthPrivileges *PrivilegesRet) map[int][]string {
	privList := make(map[int][]string)
	for _, CAuthPriv := range CAuthPrivileges.Data {
		params := strings.Split(CAuthPriv.PrivilegeName, "-")
		if len(params) != 2 {
			util.LogInfo.Printf("Error priv in converting: %s", CAuthPriv.PrivilegeName)
			continue
		}

		cmd := params[0]
		module := params[1]

		if cmd == "add" {
			cmd = "create"
		}

		if cmd == "modify" {
			cmd = "edit"
		}

		if _, ok := PrivilegesMap[module]; !ok {
			util.LogInfo.Printf("Error cauth module which not existed in system: %s\n", module)
			continue
		}

		id := PrivilegesMap[module].ID
		privList[id] = append(privList[id], cmd)
	}
	return privList
}

func convertAPIPrivToCAuthPriv(priv map[int][]string) []string {
	ret := []string{}

	idNameMap := make(map[int]string)
	idPrivMap := make(map[int]*Privilege)
	for key, val := range PrivilegesMap {
		idNameMap[val.ID] = key
		idPrivMap[val.ID] = val
	}

	for key, val := range priv {
		name := idNameMap[key]

		if len(val) == 0 {
			allCmd := strings.Split(idPrivMap[key].CmdList, ",")
			for _, cmd := range allCmd {
				if cmd == "edit" {
					cmd = "modify"
				} else if cmd == "create" {
					cmd = "add"
				}
				ret = append(ret, fmt.Sprintf("%s-%s", cmd, name))
			}
		} else {
			for _, cmd := range val {
				if cmd == "edit" {
					cmd = "modify"
				} else if cmd == "create" {
					cmd = "add"
				}
				ret = append(ret, fmt.Sprintf("%s-%s", cmd, name))
			}
		}
	}
	return ret
}

func addRole(roleName string, requestor string) error {
	param := RoleInput{
		RoleDesc:        roleName,
		RoleName:        roleName,
		ApplicationName: applicationName,
		Requestor:       requestor,
		AppKey:          getCAuthAppKey(),
	}

	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), createRoleEntry)
	ret, body, err := util.HTTPPostJSONWithStatus(postURL, param, 5)
	if err != nil {
		return err
	}
	if ret != iris.StatusOK {
		apiRet := CAuthStatus{}
		err := json.Unmarshal([]byte(body), &apiRet)
		if err != nil {
			return fmt.Errorf("Add Role fail: %d [%s]", ret, body)
		} else {
			return fmt.Errorf("Add Role fail: %s", apiRet.Error.Message)
		}
	}
	return nil
}

func deleteRole(roleName string, requestor string) error {
	param := DeleteRoleInput{
		RoleName:        roleName,
		ApplicationName: applicationName,
		Requestor:       requestor,
		AppKey:          getCAuthAppKey(),
	}

	postURL := fmt.Sprintf("%s/%s/%s", getCAuthServer(), getCAuthPrefix(), deleteRoleEntry)
	ret, body, err := util.HTTPRequestJSONWithStatus(postURL, param, 5, "DELETE")
	if err != nil {
		return err
	}
	if ret != iris.StatusOK {
		apiRet := CAuthStatus{}
		err := json.Unmarshal([]byte(body), &apiRet)
		if err != nil {
			return fmt.Errorf("Delete Role fail: %d [%s]", ret, body)
		} else {
			return fmt.Errorf("Delete Role fail: %s", apiRet.Error.Message)
		}
	}
	return nil
}
