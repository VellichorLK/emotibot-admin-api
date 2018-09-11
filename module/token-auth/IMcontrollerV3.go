package main

import (
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
	"github.com/gorilla/mux"
)

func IMUserAddHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	var user *data.UserDetailV3
	var err error

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
	user.Products = []string{"IM"}

	if user.UserName == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "username")
		return
	}
	if user.Email == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "email")
		return
	}
	if user.DisplayName == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "name")
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

	returnBFSuccess(w, newUser)
}

func IMUserUpdateHandlerV3(w http.ResponseWriter, r *http.Request) {
	requester := getRequesterV3(r)
	vars := mux.Vars(r)

	var userID string
	var origUser *data.UserDetailV3
	var err error

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

	newUser.UserName = origUser.UserName
	if newUser.Email == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "email")
		return
	}
	if newUser.DisplayName == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "name")
		return
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
		case enum.SuperAdminUser:
			fallthrough
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

	returnBFSuccess(w, true)
}

func IMValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		params := strings.Split(r.Header.Get("Authorization"), " ")
		if len(params) >= 2 {
			token = params[1]
		}
	}
	if token == "" {
		returnBadRequest(w, "token")
		return
	}

	userInfo := data.User{}
	err := userInfo.SetValueWithToken(token)
	if err != nil {
		util.LogInfo.Println("Check token fail: ", err.Error())
		returnBadRequest(w, "token")
		return
	}
	returnBFSuccess(w, userInfo)
}

func IMAppsGetHandlerV3(w http.ResponseWriter, r *http.Request) {
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

	bfData := make([]*data.BFAppV3, len(retData))
	for idx := range retData {
		bfData[idx] = &data.BFAppV3{}
		bfData[idx].CopyFromApp(retData[idx])
	}

	returnBFSuccess(w, bfData)
}
