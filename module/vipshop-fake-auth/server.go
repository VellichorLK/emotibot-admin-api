package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

func main() {
	app := iris.New()
	app.Handle("POST", "/roleRest/getAllRolesByAppName", getAllRolesByAppName)

	//adminGroup for
	adminGroup := app.Party("/", checkRequestorMiddleware)
	adminGroup.Handle("POST", "/roleRest/createRole", createRole)
	adminGroup.Handle("DELETE", "/roleRest/deleteRole", deleteRole)
	adminGroup.Handle("POST", "/rolePrivilegeRest/addRolePrivilege", addRolePrivilege)
	adminGroup.Handle("DELETE", "/rolePrivilegeRest/delRolePrivilege", delRolePrivilege)
	adminGroup.Handle("POST", "/userRoleRest/addUserRole", addUserRole)
	adminGroup.Handle("DELETE", "/userRoleRest/delUserRole", delUserRole)

	app.Handle("POST", "/privilegeRest/getPrivilegesByRole", getPrivilegesByRole)
	app.Handle("POST", "/userRoleRest/getRolesByUsers", getRolesByUsers)
	app.Handle("POST", "/userRest/getUsesByRole", getUsesByRole)

	app.Run(iris.Addr(":8787"), iris.WithoutVersionChecker)
}

func checkRequestorMiddleware(ctx context.Context) {
	//anonymous struct for reading requestor from body
	var requestor struct {
		Requestor string `json:"requestor"`
	}

	// Because ctx.ReadJSON will empty the Body, so we need to copy it out first or ReadJSON twice will emit error
	buf, _ := ioutil.ReadAll(ctx.Request().Body)
	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	err := ctx.ReadJSON(&requestor)
	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println(err)
		retError(ctx, err)
		ctx.Skip()
	}

	// Only admin user user1 can continue the admin operation
	if strings.Compare(requestor.Requestor, "user1") == 0 || strings.Compare(requestor.Requestor, "xianghong01.wang") == 0 || strings.Compare(requestor.Requestor, "richard.fu") == 0 {
		ctx.Next()
	} else {
		// Error message based on real error log
		err := fmt.Errorf("Not admin user: %s of application : 371", requestor.Requestor)
		retError(ctx, err)
		ctx.StopExecution()
	}

}

func retError(ctx context.Context, err error) {
	ctx.StatusCode(iris.StatusBadRequest)
	ctx.JSON(ErrorStatus{
		Error: ReturnStatus{
			ResponseCode: iris.StatusBadRequest,
			Message:      err.Error(),
		},
	})
}

func retSuccess(ctx context.Context) {
	ctx.JSON(SuccessStatus{
		Data: ReturnStatus{
			ResponseCode: iris.StatusOK,
			Message:      "success",
		},
	})
}

func getAllRolesByAppName(ctx context.Context) {
	input := RolesParam{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		fmt.Printf("Err: %#v\n", err)
		return
	}

	ret := []*RoleRet{}
	for _, role := range Roles {
		temp := RoleRet{
			RoleName:        role.RoleName,
			ApplicationName: role.ApplicationName,
			CreateTime:      role.CreateTime,
			LastModifyTime:  role.LastModifyTime,
			RoleDesc:        role.RoleDesc,
			RoleState:       role.RoleState,
		}
		ret = append(ret, &temp)
	}
	ctx.JSON(AllRolesRet{
		Data: ret,
	})
	fmt.Printf("Ret: %#v\n", ret)
}

func createRole(ctx context.Context) {
	input := RoleInput{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	if _, ok := Roles[input.RoleName]; ok {
		retError(ctx, errors.New("Duplicated Role"))
		return
	}

	now := time.Now().Unix()
	newRole := StoreRole{
		RoleName:        input.RoleName,
		ApplicationName: input.ApplicationName,
		CreateTime:      now,
		LastModifyTime:  now,
		RoleDesc:        input.RoleDesc,
		RoleState:       1,
		Privileges:      []string{},
	}
	Roles[input.RoleName] = &newRole
	retSuccess(ctx)
}

func deleteRole(ctx context.Context) {
	input := RoleInput{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	delete(Roles, input.RoleName)
	retSuccess(ctx)
}

func getPrivilegesByRole(ctx context.Context) {
	input := RolePrivilegesParam{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	ret := []*PrivilegeRet{}
	if role, ok := Roles[input.RoleName]; ok {
		fmt.Printf("Get role: %+v\n", role)
		for _, privStr := range role.Privileges {
			if priv, ok := Privileges[privStr]; ok {
				ret = append(ret, &PrivilegeRet{
					PrivilegeName: priv.PrivilegeName,
					AssetName:     priv.AssetName,
				})
			} else {
				retError(ctx, errors.New("Not existed privilege"))
				return
			}
		}
	} else {
		retError(ctx, errors.New("Invalid roleName"))
		return
	}
	ctx.JSON(PrivilegesRet{
		Data: ret,
	})
}

func addRolePrivilege(ctx context.Context) {
	input := RolePrivilegeInput{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	fmt.Printf("Get input: %#v\n", input)
	if role, ok := Roles[input.RoleName]; ok {
		if _, ok := Privileges[input.PrivilegeName]; !ok {
			retError(ctx, errors.New("privName invalid"))
			return
		}
		fmt.Printf("Try to add: %s\n", input.PrivilegeName)
		if !Contains(role.Privileges, input.PrivilegeName) {
			role.Privileges = append(role.Privileges, input.PrivilegeName)
			fmt.Printf("Add success: %s, %#v\n", input.PrivilegeName, role.Privileges)
		}
	} else {
		retError(ctx, errors.New("roleName invalid"))
		return
	}

	retSuccess(ctx)
}

func delRolePrivilege(ctx context.Context) {
	input := RolePrivilegeInput{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	if role, ok := Roles[input.RoleName]; ok {
		if _, ok := Privileges[input.PrivilegeName]; !ok {
			fmt.Printf("privName invalid")
			retError(ctx, errors.New("privName invalid"))
			return
		}
		role.Privileges = Remove(role.Privileges, input.PrivilegeName)
		fmt.Printf("Remove role %s: %+v", input.PrivilegeName, role)
	} else {
		fmt.Printf("roleName invalid")
		retError(ctx, errors.New("roleName invalid"))
		return
	}

	retSuccess(ctx)
}

func getRolesByUsers(ctx context.Context) {
	input := UserRolesParam{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	ret := make(map[string][]*SimpleRoleRet)
	for _, userAccount := range input.UserAccounts {
		if user, ok := Users[userAccount]; ok {
			roles := user.Roles
			if len(user.Roles) <= 0 {
				retError(ctx, errors.New("Role must not empty"))
				return
			}

			ret[userAccount] = []*SimpleRoleRet{}
			for _, roleStr := range roles {
				if role, ok := Roles[roleStr]; ok {
					ret[userAccount] = append(ret[userAccount], &SimpleRoleRet{
						RoleName:       role.RoleName,
						CreateTime:     role.CreateTime,
						LastModifyTime: role.LastModifyTime,
						RoleDesc:       role.RoleDesc,
					})
				} else {
					retError(ctx, errors.New("Role name not existed"))
					return
				}
			}
		} else {
			retError(ctx, errors.New("Account not available"))
			return
		}
	}
	ctx.JSON(UserRolesRet{
		Data: ret,
	})
}

func addUserRole(ctx context.Context) {
	input := UserRoleInput{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	if _, ok := Roles[input.RoleName]; !ok {
		retError(ctx, errors.New("roleName invalid"))
		return
	}

	if user, ok := Users[input.UserAccount]; ok {
		if !Contains(user.Roles, input.RoleName) {
			user.Roles = append(user.Roles, input.RoleName)
		}
	} else {
		retError(ctx, fmt.Errorf("User %s is not existed", input.UserAccount))
		return
	}

	retSuccess(ctx)
}

func delUserRole(ctx context.Context) {
	input := UserRoleInput{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	if user, ok := Users[input.UserAccount]; ok {
		if _, ok := Roles[input.RoleName]; !ok {
			retError(ctx, errors.New("roleName invalid"))
			return
		}
		user.Roles = Remove(user.Roles, input.RoleName)
	} else {
		retError(ctx, errors.New("userAccount invalid"))
		return
	}

	retSuccess(ctx)
}

func getUsesByRole(ctx context.Context) {
	input := RoleUsersParam{}
	err := ctx.ReadJSON(&input)
	if err != nil {
		retError(ctx, err)
		return
	}

	ret := []*UserRet{}
	for _, user := range Users {
		if Contains(user.Roles, input.RoleName) {
			ret = append(ret, &UserRet{
				UserName:       user.UserName,
				UserDepartment: user.UserDepartment,
				UserAccountID:  user.UserAccountID,
				UserCode:       user.UserCode,
			})
		}
	}
	ctx.JSON(UsersRet{
		Data: ret,
	})
}
