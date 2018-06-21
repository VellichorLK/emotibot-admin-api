package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/audit"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
	"github.com/gorilla/mux"
)

func SystemAdminsGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	retData, err := service.GetSystemAdminsV3()
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, retData)
}

func SystemAdminGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	adminID := vars["adminID"]
	if !util.IsValidUUID(adminID) {
		returnBadRequest(w, "admin-id")
		return
	}

	retData, err := service.GetSystemAdminV3(adminID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func SystemAdminAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	admin, err := parseAddUserFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}
	admin.Type = enum.SuperAdminUser

	id, err := service.AddSystemAdminV3(admin)
	if err != nil {
		switch err {
		case util.ErrUserNameExists:
			returnBadRequest(w, "username")
		case util.ErrUserEmailExists:
			returnBadRequest(w, "email")
		default:
			returnInternalError(w, err.Error())
		}
		return
	}

	newAdmin, err := service.GetSystemAdminV3(id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, newAdmin)
}

func SystemAdminUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	adminID := vars["adminID"]
	if !util.IsValidUUID(adminID) {
		returnBadRequest(w, "admin-id")
		return
	}

	origAdmin, err := service.GetSystemAdminV3(adminID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origAdmin == nil {
		returnNotFound(w)
		return
	}

	newAdmin, err := parseUpdateUserFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}
	newAdmin.Type = enum.SuperAdminUser

	if newAdmin.UserName == "" {
		newAdmin.UserName = origAdmin.UserName
	}
	if newAdmin.DisplayName == "" {
		newAdmin.DisplayName = origAdmin.DisplayName
	}
	if newAdmin.Email == "" {
		newAdmin.Email = origAdmin.Email
	}
	if newAdmin.Phone == "" {
		newAdmin.Phone = origAdmin.Phone
	}

	if *newAdmin.Password != "" {
		verifyPassword := r.FormValue("verify_password")

		if verifyPassword == "" {
			returnForbidden(w)
			return
		}

		password, err := service.GetUserPasswordV3(requester.ID)
		if err != nil {
			returnInternalError(w, err.Error())
			return
		} else if password == "" {
			returnForbidden(w)
			return
		}

		if verifyPassword != password {
			returnForbidden(w)
			return
		}
	}

	err = service.UpdateSystemAdminV3(origAdmin, newAdmin, adminID)
	if err != nil {
		switch err {
		case util.ErrUserNameExists:
			returnBadRequest(w, "username")
			return
		case util.ErrUserEmailExists:
			returnBadRequest(w, "email")
			return
		default:
			returnInternalError(w, err.Error())
			return
		}
	}

	returnSuccess(w, true)
}

func SystemAdminDeleteHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	if requester.Type != enum.SuperAdminUser {
		returnForbidden(w)
		return
	}

	adminID := vars["adminID"]
	if !util.IsValidUUID(adminID) {
		returnBadRequest(w, "admin-id")
		return
	}

	result, err := service.DeleteSystemAdminV3(adminID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if !result {
		returnNotFound(w)
		return
	}

	returnSuccess(w, true)
}

func EnterprisesGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	retData, err := service.GetEnterprisesV3()
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, retData)
}

func EnterpriseGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	retData, err := service.GetEnterpriseV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func EnterpriseAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		returnBadRequest(w, "name")
		return
	}

	adminUser := r.FormValue("admin")
	if adminUser == "" {
		returnBadRequest(w, "admin")
		return
	}

	adminReq := data.EnterpriseAdminRequestV3{}
	err := json.Unmarshal([]byte(adminUser), &adminReq)
	if err != nil {
		returnBadRequest(w, "admin")
		return
	}

	enterpriseAdmin := data.UserDetailV3{
		UserV3: data.UserV3{
			UserName:    adminReq.Account,
			DisplayName: adminReq.Name,
			Type:        enum.AdminUser,
		},
		Password: &adminReq.Password,
	}

	description := r.FormValue("description")

	var modules []string
	err = json.Unmarshal([]byte(r.FormValue("modules")), &modules)
	if err != nil {
		returnBadRequest(w, "modules")
		return
	}

	enterprise := data.EnterpriseV3{
		Name:        name,
		Description: description,
	}

	id, err := service.AddEnterpriseV3(&enterprise, modules, &enterpriseAdmin)
	if err != nil {
		switch err {
		case util.ErrEnterpriseInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	}

	newEnterprise, err := service.GetEnterpriseV3(id)
	if err != nil || newEnterprise == nil {
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, newEnterprise)
}

func EnterpriseUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	origEnterprise, err := service.GetEnterpriseV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origEnterprise == nil {
		returnNotFound(w)
		return
	}

	newEnterprise, err := parseEnterpriseFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if newEnterprise.Name == "" {
		newEnterprise.Name = origEnterprise.Name
	}
	if newEnterprise.Description == "" {
		newEnterprise.Description = origEnterprise.Description
	}

	modules := strings.Split(r.FormValue("modules"), ",")

	err = service.UpdateEnterpriseV3(enterpriseID, origEnterprise, newEnterprise, modules)
	if err != nil {
		switch err {
		case util.ErrEnterpriseInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	}

	returnSuccess(w, true)
}

func EnterpriseDeleteHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	result, err := service.DeleteEnterpriseV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if !result {
		returnNotFound(w)
		return
	}

	returnSuccess(w, true)
}

func UsersGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	retData, err := service.GetUsersV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func UserGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	userID := vars["userID"]
	if !util.IsValidUUID(userID) {
		returnBadRequest(w, "user-id")
		return
	}

	retData, err := service.GetUserV3(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func UserAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	var user *data.UserDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if user != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentUserAdd, user.UserName)
		} else {
			auditMessage = data.AuditContentUserAdd
		}

		addAuditLog(r, audit.AuditModuleMembers, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	user, err = parseAddUserFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if requester.Type > user.Type {
		err = util.ErrOperationForbidden
		returnForbidden(w)
		return
	}

	id, err := service.AddUserV3(enterpriseID, user)
	if err != nil {
		switch err {
		case util.ErrRobotGroupNotExist:
			fallthrough
		case util.ErrRobotNotExist:
			fallthrough
		case util.ErrRoleNotExist:
			returnUnprocessableEntity(w, err.Error())
		case util.ErrUserNameExists:
			returnBadRequest(w, "username")
		case util.ErrUserEmailExists:
			returnBadRequest(w, "email")
		default:
			returnInternalError(w, err.Error())
		}
		return
	} else if id == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	newUser, err := service.GetUserV3(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if newUser == nil {
		err = util.ErrInteralServer
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, newUser)
}

func UserUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	var userID string
	var origUser *data.UserDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if origUser != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentUserUpdate, origUser.UserName)
		} else if userID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentUserUpdate, userID)
		} else {
			auditMessage = data.AuditContentUserUpdate
		}

		addAuditLog(r, audit.AuditModuleMembers, audit.AuditOperationEdit,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	userID = vars["userID"]
	if !util.IsValidUUID(userID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "user-id")
		return
	}

	origUser, err = service.GetUserV3(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origUser == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	if requester.Type > origUser.Type {
		err = util.ErrOperationForbidden
		returnForbidden(w)
		return
	}

	newUser, err := parseUpdateUserFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if newUser.UserName == "" {
		newUser.UserName = origUser.UserName
	}
	if newUser.DisplayName == "" {
		newUser.DisplayName = origUser.DisplayName
	}
	if newUser.Email == "" {
		newUser.Email = origUser.Email
	}
	if newUser.Phone == "" {
		newUser.Phone = origUser.Phone
	}

	if *newUser.Password != "" {
		verifyPassword := r.FormValue("verify_password")

		if verifyPassword == "" {
			err = util.ErrOperationForbidden
			returnForbidden(w)
			return
		}

		var password string

		switch requester.Type {
		case enum.AdminUser:
			password, err = service.GetUserPasswordV3(requester.ID)
			if err != nil {
				returnInternalError(w, err.Error())
				return
			} else if password == "" {
				err = util.ErrOperationForbidden
				returnForbidden(w)
				return
			}
		case enum.NormalUser:
			password = *origUser.Password
		default:
			err = util.ErrOperationForbidden
			returnForbidden(w)
			return
		}

		if verifyPassword != password {
			err = util.ErrOperationForbidden
			returnForbidden(w)
			return
		}
	}

	err = service.UpdateUserV3(enterpriseID, userID, origUser, newUser)
	if err != nil {
		switch err {
		case util.ErrRobotGroupNotExist:
			fallthrough
		case util.ErrRobotNotExist:
			fallthrough
		case util.ErrRoleNotExist:
			returnUnprocessableEntity(w, err.Error())
			return
		case util.ErrUserNameExists:
			returnBadRequest(w, "username")
			return
		case util.ErrUserEmailExists:
			returnBadRequest(w, "email")
			return
		default:
			returnInternalError(w, err.Error())
			return
		}
	}

	returnSuccess(w, true)
}

func UserDeleteHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	var userID string
	var user *data.UserDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if user != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentUserDelete, user.UserName)
		} else if userID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentUserDelete, userID)
		} else {
			auditMessage = data.AuditContentUserDelete
		}

		addAuditLog(r, audit.AuditModuleMembers, audit.AuditOperationDelete,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}
	userID = vars["userID"]
	if !util.IsValidUUID(userID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "user-id")
		return
	}

	user, err = service.GetUserV3(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if user == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	if requester.Type > user.Type {
		err = util.ErrOperationForbidden
		returnForbidden(w)
		return
	}

	result, err := service.DeleteUserV3(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if !result {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	returnSuccess(w, true)
}

func AppsGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	retData, err := service.GetAppsV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
	}

	returnSuccess(w, retData)
}

func AppGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	appID := vars["appID"]
	if !util.IsValidUUID(appID) {
		returnBadRequest(w, "app-id")
		return
	}

	retData, err := service.GetAppV3(enterpriseID, appID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func AppAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var app *data.AppDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if app != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppAdd, app.Name)
		} else {
			auditMessage = data.AuditContentAppAdd
		}

		addAuditLog(r, audit.AuditModuleRobot, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	app, err = parseAppFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	id, err := service.AddAppV3(enterpriseID, app)
	if err != nil {
		switch err {
		case util.ErrAppInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	} else if id == "" {
		returnBadRequest(w, "enterprise-id")
		return
	}

	newApp, err := service.GetAppV3(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if newApp == nil {
		err = util.ErrInteralServer
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, newApp)
}

func AppUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var appID string
	var origApp *data.AppDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if origApp != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppUpdate, origApp.Name)
		} else if appID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppUpdate, appID)
		} else {
			auditMessage = data.AuditContentAppUpdate
		}

		addAuditLog(r, audit.AuditModuleRobot, audit.AuditOperationEdit,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	appID = vars["appID"]
	if !util.IsValidUUID(appID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "app-id")
		return
	}

	origApp, err = service.GetAppV3(enterpriseID, appID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origApp == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	newApp, err := parseAppFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if newApp.Name == "" {
		newApp.Name = origApp.Name
	}
	if newApp.Description == "" {
		newApp.Description = origApp.Description
	}

	err = service.UpdateAppV3(enterpriseID, appID, origApp, newApp)
	if err != nil {
		switch err {
		case util.ErrAppInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	}

	returnSuccess(w, true)
}

func AppDeleteHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var appID string
	var app *data.AppDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if app != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppDelete, app.Name)
		} else if appID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppDelete, appID)
		} else {
			auditMessage = data.AuditContentAppDelete
		}

		addAuditLog(r, audit.AuditModuleRobot, audit.AuditOperationDelete,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	appID = vars["appID"]
	if !util.IsValidUUID(appID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "app-id")
		return
	}

	app, err = service.GetAppV3(enterpriseID, appID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if app == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	result, err := service.DeleteAppV3(enterpriseID, appID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if !result {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	returnSuccess(w, true)
}

func GroupsGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	retData, err := service.GetGroupsV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func GroupGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	groupID := vars["groupID"]
	if !util.IsValidUUID(groupID) {
		returnBadRequest(w, "group-id")
		return
	}

	retData, err := service.GetGroupV3(enterpriseID, groupID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func GroupAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var group *data.GroupDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if group != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentGroupAdd, group.Name)
		} else {
			auditMessage = data.AuditContentGroupAdd
		}

		addAuditLog(r, audit.AuditModuleRobotGroup, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	group, apps, err := parseGroupFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	id, err := service.AddGroupV3(enterpriseID, group, apps)
	if err != nil {
		switch err {
		case util.ErrGroupInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	} else if id == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	newGroup, err := service.GetGroupV3(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if newGroup == nil {
		err = util.ErrInteralServer
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, newGroup)
}

func GroupUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var groupID string
	var origGroup *data.GroupDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if origGroup != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentGroupUpdate, origGroup.Name)
		} else if groupID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentGroupUpdate, groupID)
		} else {
			auditMessage = data.AuditContentGroupUpdate
		}

		addAuditLog(r, audit.AuditModuleRobotGroup, audit.AuditOperationEdit,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	groupID = vars["groupID"]
	if !util.IsValidUUID(groupID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "group-id")
		return
	}

	origGroup, err = service.GetGroupV3(enterpriseID, groupID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origGroup == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	newGroup, apps, err := parseGroupFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if newGroup.Name == "" {
		newGroup.Name = origGroup.Name
	}

	err = service.UpdateGroupV3(enterpriseID, groupID, origGroup, newGroup, apps)
	if err != nil {
		switch err {
		case util.ErrGroupInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	}

	returnSuccess(w, true)
}

func GroupDeleteHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var groupID string
	var group *data.GroupDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if group != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentGroupDelete, group.Name)
		} else if groupID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentGroupDelete, groupID)
		} else {
			auditMessage = data.AuditContentGroupDelete
		}

		addAuditLog(r, audit.AuditModuleRobotGroup, audit.AuditOperationDelete,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	groupID = vars["groupID"]
	if !util.IsValidUUID(groupID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "group-id")
		return
	}

	group, err = service.GetGroupV3(enterpriseID, groupID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if group == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	result, err := service.DeleteGroupV3(enterpriseID, groupID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if !result {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	returnSuccess(w, true)
}

func RolesGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, err := service.GetRolesV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func RoleGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	roleID := vars["roleID"]
	if !util.IsValidUUID(roleID) {
		returnBadRequest(w, "role-id")
		return
	}

	retData, err := service.GetRoleV3(enterpriseID, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func RoleAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var role *data.RoleV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if role != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentRoleAdd, role.Name)
		} else {
			auditMessage = data.AuditContentRoleAdd
		}

		addAuditLog(r, audit.AuditModuleRole, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	role, err = parseRoleFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	id, err := service.AddRoleV3(enterpriseID, role)
	if err != nil {
		switch err {
		case util.ErrRoleInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	} else if id == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	newRole, err := service.GetRoleV3(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if newRole == nil {
		err = util.ErrInteralServer
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, newRole)
}

func RoleUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var roleID string
	var origRole *data.RoleV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if origRole != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentRoleUpdate, origRole.Name)
		} else if roleID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentRoleUpdate, roleID)
		} else {
			auditMessage = data.AuditContentRoleUpdate
		}

		addAuditLog(r, audit.AuditModuleRole, audit.AuditOperationEdit,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	roleID = vars["roleID"]
	if !util.IsValidUUID(roleID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "role-id")
		return
	}

	origRole, err = service.GetRoleV3(enterpriseID, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origRole == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	newRole, err := parseRoleFromRequestV3(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	err = service.UpdateRoleV3(enterpriseID, roleID, origRole, newRole)
	if err != nil {
		switch err {
		case util.ErrRoleInfoExists:
			returnBadRequest(w, "name")
		default:
			returnInternalError(w, err.Error())
		}
		return
	}

	returnSuccess(w, true)
}

func RoleDeleteHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var roleID string
	var role *data.RoleV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if role != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentRoleDelete, role.Name)
		} else if roleID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentRoleDelete, roleID)
		} else {
			auditMessage = data.AuditContentRoleDelete
		}

		addAuditLog(r, audit.AuditModuleRole, audit.AuditOperationDelete,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	roleID = vars["roleID"]
	if !util.IsValidUUID(roleID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "role-id")
		return
	}

	role, err = service.GetRoleV3(enterpriseID, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if role == nil {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	result, err := service.DeleteRoleV3(enterpriseID, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else if !result {
		err = util.ErrResourceNotFound
		returnNotFound(w)
		return
	}

	returnSuccess(w, true)
}

func LoginHandlerV3(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	account := r.Form.Get("account")
	passwd := r.Form.Get("passwd")

	var err error

	defer func() {
		// Add audit log
		auditMessage := fmt.Sprintf("[%s]: %s", data.AuditLogin, account)
		addAuditLog(r, audit.AuditModuleMembers, audit.AuditOperationLogin,
			auditMessage, err)
	}()

	if !util.IsValidString(&passwd) || !util.IsValidString(&account) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "")
		return
	}

	// If user is banned, return Forbidden
	if util.UserBanInfos.IsUserBanned(account) {
		err = util.ErrOperationForbidden
		returnForbidden(w)
		writeErrJSONWithObj(w, "forbidden", util.UserBanInfos[account])
		return
	}

	user, err := service.LoginV3(account, passwd)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if user == nil {
		// Login fail
		addUserTryCount(account)
		fmt.Printf("User %s login fail: %d\n", account, userTryCount[account])
		// Ban user if it's retry time more than 5
		if getUserTryCount(account) > banRetryTimes {
			util.UserBanInfos.BanUser(account)
			resetUserTryCount(account)
		}
		err = util.ErrOperationForbidden
		returnForbidden(w)
		writeErrJSONWithObj(w, "forbidden", util.UserBanInfos[account])
		return
	}

	// Login success, clear ban info
	util.UserBanInfos.ClearBanInfo(account)

	token, err := user.GenerateToken()
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	ret := data.LoginInfoV3{
		Token: token,
		Info:  user,
	}

	returnSuccess(w, ret)
}

func ModulesGetHandlerV3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, err := service.GetModulesV3(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func parseEnterpriseFromRequestV3(r *http.Request) (*data.EnterpriseDetailV3, error) {
	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		return nil, errors.New("invalid name")
	}

	enterprise := data.EnterpriseDetailV3{}
	enterprise.Name = name
	enterprise.Description = description

	return &enterprise, nil
}

func parseAddUserFromRequestV3(r *http.Request) (*data.UserDetailV3, error) {
	user := loadUserFromRequestV3(r)

	// if user.Email == nil || *user.Email == "" {
	// 	return nil, errors.New("invalid email")
	// }
	if user.Password == nil || *user.Password == "" {
		return nil, errors.New("Invalid password")
	}
	if user.UserName == "" {
		return nil, errors.New("Invalid username")
	}

	return user, nil
}

func parseUpdateUserFromRequestV3(r *http.Request) (*data.UserDetailV3, error) {
	user := loadUserFromRequestV3(r)

	// if user.Email == nil || *user.Email == "" {
	// 	return nil, errors.New("invalid email")
	// }

	return user, nil
}

func loadUserFromRequestV3(r *http.Request) *data.UserDetailV3 {
	username := r.FormValue("username")
	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	password := r.FormValue("password")

	user := data.UserDetailV3{}
	user.Email = email
	user.Phone = phone
	user.DisplayName = name
	user.Password = &password
	user.UserName = username

	userType, err := strconv.Atoi(r.FormValue("type"))
	if err != nil {
		userType = enum.NormalUser
	} else if userType > enum.NormalUser || userType < enum.AdminUser {
		userType = enum.NormalUser
	}
	user.Type = userType

	roles := r.FormValue("apps")
	if roles != "" {
		userRolesReq := data.UserRolesRequestV3{}
		err = json.Unmarshal([]byte(roles), &userRolesReq)
		if err == nil {
			userRoles := data.UserRolesV3{
				GroupRoles: make([]*data.UserGroupRoleV3, 0),
				AppRoles:   make([]*data.UserAppRoleV3, 0),
			}

			for group, roles := range userRolesReq.GroupRoles {
				for _, role := range roles {
					userGroup := data.UserGroupRoleV3{
						ID:   group,
						Role: role,
					}
					userRoles.GroupRoles = append(userRoles.GroupRoles, &userGroup)
				}
			}

			for app, roles := range userRolesReq.AppRoles {
				for _, role := range roles {
					userApp := data.UserAppRoleV3{
						ID:   app,
						Role: role,
					}
					userRoles.AppRoles = append(userRoles.AppRoles, &userApp)
				}
			}

			user.Roles = &userRoles
		} else {
			util.LogTrace.Println("Parse json error: ", err.Error())
		}
	}

	customStr := r.FormValue("custom")
	if customStr != "" {
		customInfo := map[string]string{}
		err = json.Unmarshal([]byte(customStr), &customInfo)
		if err == nil {
			user.CustomInfo = &customInfo
		} else {
			util.LogTrace.Println("Parse json error: ", err.Error())
		}
	}

	return &user
}

func parseAppFromRequestV3(r *http.Request) (*data.AppDetailV3, error) {
	name := r.FormValue("name")
	description := r.FormValue("description")

	ret := data.AppDetailV3{
		AppV3: data.AppV3{
			Name: name,
		},
		Description: description,
	}

	return &ret, nil
}

func parseGroupFromRequestV3(r *http.Request) (*data.GroupDetailV3, []string, error) {
	name := r.FormValue("name")

	var apps []string
	err := json.Unmarshal([]byte(r.FormValue("apps")), &apps)
	if err != nil {
		return nil, nil, err
	}

	group := data.GroupDetailV3{
		GroupV3: data.GroupV3{
			Name: name,
		},
	}

	return &group, apps, nil
}

func parseRoleFromRequestV3(r *http.Request) (*data.RoleV3, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, errors.New("Invalid name")
	}
	description := r.FormValue("description")
	privilegeStr := r.FormValue("privilege")

	privileges := map[string][]string{}
	err := json.Unmarshal([]byte(privilegeStr), &privileges)
	if err != nil {
		util.LogError.Printf("Cannot decode privilegeStr: %s\n", err.Error())
		return nil, err
	}

	ret := data.RoleV3{
		Name:        name,
		Description: description,
		Privileges:  privileges,
	}

	return &ret, nil
}

func addAuditLog(r *http.Request, auditModule, auditOperation,
	auditMessage string, err error) {
	appID := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

	result := 1
	if err != nil {
		result = 0
	}

	audit.AddAuditLog(appID, userID, userIP, auditModule, auditOperation,
		auditMessage, result)
}