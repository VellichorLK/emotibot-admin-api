package CAuth

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

const validAppID = "vipshop"

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cauth",
		EntryPoints: []util.EntryPoint{
			// Privileges and users is readonly in VIP's CAuth system
			util.NewEntryPoint("GET", "privileges", []string{}, handleListPrivilege),
			util.NewEntryPoint("GET", "users", []string{}, handleUserList),
			util.NewEntryPoint("GET", "roles", []string{}, handleRoleList),

			util.NewEntryPoint("GET", "user/{id:string}", []string{}, handleUserGet),
			util.NewEntryPoint("PATCH", "user/{id:string}", []string{}, handleUserUpdate),
			util.NewEntryPoint("PATCH", "role/{id:string}", []string{}, handleRoleUpdate),

			util.NewEntryPoint("POST", "role/register", []string{}, handleAddRole),

			util.NewEntryPoint("DELETE", "role/{id:string}", []string{}, handleDeleteRole),
		},
	}
}

func handleListPrivilege(ctx context.Context) {
	appid := util.GetAppID(ctx)
	if appid != validAppID {
		ctx.StatusCode(iris.StatusUnauthorized)
		return
	}

	ret := []*Privilege{}
	for _, priv := range PrivilegesMap {
		ret = append(ret, priv)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ID < ret[j].ID
	})

	ctx.JSON(GenRetObj(ApiError.SUCCESS, ret))
}

func handleRoleList(ctx context.Context) {
	appid := util.GetAppID(ctx)
	if appid != validAppID {
		ctx.StatusCode(iris.StatusUnauthorized)
		return
	}

	CAuthRoles, err := getRolesFromCAuth()
	if err != nil {
		ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err))
		return
	}

	ret := []*Role{}
	for _, role := range CAuthRoles.Data {
		temp := &Role{
			RoleID:   role.RoleName,
			RoleName: role.RoleName,
		}

		privList, err := GetRolePrivs(role.RoleName)
		if err != nil {
			ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
			return
		}
		privStr, _ := json.Marshal(privList)
		temp.Privilege = string(privStr)

		ret = append(ret, temp)
	}

	ctx.JSON(GenRetObj(ApiError.SUCCESS, ret))
}

func handleUserList(ctx context.Context) {
	appid := util.GetAppID(ctx)
	if appid != validAppID {
		ctx.StatusCode(iris.StatusUnauthorized)
		return
	}

	CAuthRoles, err := getRolesFromCAuth()
	if err != nil {
		ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		return
	}

	userIDList := []string{}
	users := []*UserProp{}
	for _, role := range CAuthRoles.Data {
		CAuthUsers, err := getUsersOfRoleFromCAuth(role.RoleName)
		if err != nil {
			ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
			return
		}
		for _, CAuthUser := range CAuthUsers.Data {
			if util.Contains(userIDList, CAuthUser.UserAcountID) {
				continue
			}
			fmt.Printf("Get user [%s]:%s\n", role.RoleName, CAuthUser.UserName)
			user := &UserProp{
				UserId:   CAuthUser.UserAcountID,
				UserName: CAuthUser.UserName,
				UserType: 1,
				RoleId:   role.RoleName,
			}
			users = append(users, user)
			userIDList = append(userIDList, CAuthUser.UserAcountID)
		}
	}

	ctx.JSON(GenRetObj(ApiError.SUCCESS, users))
}

func handleUserGet(ctx context.Context) {
	id := ctx.Params().GetEscape("id")
	appid := util.GetAppID(ctx)
	if appid != validAppID {
		ctx.StatusCode(iris.StatusUnauthorized)
		return
	}

	user := &SingleUserProp{
		UserId:   id,
		UserType: 1,
	}
	roles, err := GetUserRoles(id)
	if err != nil {
		ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		return
	}

	if len(roles) <= 0 {
		user.RoleId = ""
		user.Privileges = "{}"
		ctx.JSON(GenRetObj(ApiError.SUCCESS, user))
		return
	}

	role := roles[0]
	privList, err := GetRolePrivs(role.RoleName)
	if err != nil {
		ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		return
	}
	privStr, _ := json.Marshal(privList)

	user.RoleId = role.RoleName
	user.Privileges = string(privStr)
	ctx.JSON(GenRetObj(ApiError.SUCCESS, user))
	return
}

func handleUserUpdate(ctx context.Context) {
	id := ctx.Params().GetEscape("id")
	appid := util.GetAppID(ctx)
	if appid != validAppID {
		ctx.StatusCode(iris.StatusUnauthorized)
		return
	}
	operator := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	result := 0

	if len(strings.Trim(id, " ")) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenSimpleRetObj(ApiError.REQUEST_ERROR))
		return
	}

	roleID := strings.Trim(ctx.FormValue("role_id"), " ")

	origUserRoles, err := GetUserRoles(id)
	if err != nil {
		util.LogTrace.Printf("Cannot get orig role of user, %s", err.Error())
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		return
	}

	origRoles := []string{}
	for _, role := range origUserRoles {
		origRoles = append(origRoles, role.RoleName)
	}

	logMsg := ""
	operation := util.AuditOperationEdit
	if roleID == "" {
		// Update role to empty means delete user in VCA application
		logMsg = fmt.Sprintf("%s%s %s", util.Msg["Delete"], util.Msg["User"], id)
		operation = util.AuditOperationDelete
	} else if len(origRoles) == 0 {
		// Update role from empty means add user in VCA application
		logMsg = fmt.Sprintf("%s%s %s", util.Msg["Add"], util.Msg["User"], id)
		operation = util.AuditOperationAdd
	} else {
		logMsg = fmt.Sprintf("%s%s (%s) %s: %s -> %s",
			util.Msg["Modify"], util.Msg["User"], id, util.Msg["Role"], strings.Join(origRoles, ","), roleID)
	}

	err = updateUserRole(operator, id, origUserRoles, roleID)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		ctx.JSON(util.GenSimpleRetObj(ApiError.SUCCESS))
		result = 1
	}
	util.AddAuditLog(operator, userIP, util.AuditModuleMembers, operation, logMsg, result)
}

func handleRoleUpdate(ctx context.Context) {
	id := ctx.Params().GetEscape("id")
	appid := util.GetAppID(ctx)
	if appid != validAppID {
		ctx.StatusCode(iris.StatusUnauthorized)
		return
	}
	operator := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	result := 0

	if len(strings.Trim(id, " ")) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenSimpleRetObj(ApiError.REQUEST_ERROR))
		return
	}

	if len(strings.Trim(operator, " ")) == 0 {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.JSON(util.GenSimpleRetObj(ApiError.REQUEST_ERROR))
		return
	}

	newPrivStr := ctx.FormValue("privilege")
	newRolePriv := make(map[int][]string)
	err := json.Unmarshal([]byte(newPrivStr), &newRolePriv)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
		return
	}

	origRolePriv, err := GetRolePrivs(id)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		return
	}

	origPrivStr, _ := json.Marshal(origRolePriv)
	// newPrivStr, _ := json.Marshal(newRolePriv)
	logMsg := fmt.Sprintf("Update role (%s) priv: [%s] -> [%s]", id, origPrivStr, newPrivStr)

	err = updateRolePriv(operator, id, origRolePriv, newRolePriv)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		ret := Role{
			Privilege: newPrivStr,
			RoleID:    id,
			RoleName:  id,
		}
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, ret))
		result = 1
	}
	util.AddAuditLog(operator, userIP, util.AuditModuleRole, util.AuditOperationEdit, logMsg, result)
}

func handleAddRole(ctx context.Context) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	roleName := strings.Trim(ctx.FormValue("role_name"), " ")
	result := 0

	err := addRole(roleName, userID)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		ret := Role{
			Privilege: "{}",
			RoleID:    roleName,
			RoleName:  roleName,
		}
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, ret))
		result = 1
	}
	logMsg := fmt.Sprintf("Add new role: %s", roleName)
	util.AddAuditLog(userID, userIP, util.AuditModuleRole, util.AuditOperationAdd, logMsg, result)
}

func handleDeleteRole(ctx context.Context) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	id := ctx.Params().GetEscape("id")
	result := 0
	logMsg := ""

	defer func() {
		util.AddAuditLog(userID, userIP, util.AuditModuleRole, util.AuditOperationDelete, logMsg, result)
	}()

	CAuthUsers, err := getUsersOfRoleFromCAuth(id)
	if err != nil {
		ctx.JSON(GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		return
	}
	if len(CAuthUsers.Data) > 0 {
		logMsg = fmt.Sprintf("%s%s%s %s: %s%s", util.Msg["Cannot"], util.Msg["Delete"], util.Msg["Role"], id, util.Msg["Has"], util.Msg["User"])
		ctx.JSON(GenRetObj(ApiError.REQUEST_ERROR, logMsg))
		return
	}

	err = deleteRole(id, userID)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		ctx.JSON(util.GenSimpleRetObj(ApiError.SUCCESS))
		logMsg = fmt.Sprintf("%s%s %s", util.Msg["Delete"], util.Msg["Role"], id)
		result = 1
	}
}
