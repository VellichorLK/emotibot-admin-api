package data

import (
	"encoding/json"

	"emotibot.com/emotigo/module/token-auth/internal/util"
)

// User store the basic logging information of user
type UserV3 struct {
	ID          string       `json:"id"`
	UserName    string       `json:"user_name"`
	DisplayName string       `json:"display_name"`
	Email       string       `json:"email"`
	Phone       string       `json:"phone"`
	Type        int          `json:"type"`
	Roles       *UserRolesV3 `json:"roles"`
}

type UserDetailV3 struct {
	UserV3
	Password   *string            `json:"-"`
	Enterprise *string            `json:"enterprise"`
	Status     int                `json:"status"`
	CustomInfo *map[string]string `json:"custom"`
}

type UserGroupRoleV3 struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type UserAppRoleV3 struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type UserRolesV3 struct {
	GroupRoles []*UserGroupRoleV3 `json:"groups"`
	AppRoles   []*UserAppRoleV3   `json:"apps"`
}

type UserRolesRequestV3 struct {
	GroupRoles map[string][]string `json:"groups"`
	AppRoles   map[string][]string `json:"apps"`
}

// IsValid will check valid of not
func (user UserDetailV3) IsValid() bool {
	// return util.IsValidString(user.Email) &&
	// 	util.IsValidString(user.Password) &&
	// 	util.IsValidMD5(*user.Password) &&
	// 	util.IsValidString(user.Enterprise) &&
	// 	util.IsValidUUID(*user.Enterprise)
	return util.IsValidString(&user.UserName) &&
		util.IsValidString(user.Password) &&
		util.IsValidMD5(*user.Password)
}

// IsActive will check user is active or not
func (user UserDetailV3) IsActive() bool {
	return user.Status == userActive
}

// GenerateToken will generate json web token for current user
func (user UserDetailV3) GenerateToken() (string, error) {
	return util.GetJWTTokenWithCustomInfo(&user)
}

// SetValueWithToken will return an userObj from custom column of token
func (user *UserDetailV3) SetValueWithToken(tokenString string) error {
	info, err := util.ResolveJWTToken(tokenString)
	if err != nil {
		return err
	}
	jsonByte, _ := json.Marshal(info)

	userInfo := UserDetailV3{}
	err = json.Unmarshal(jsonByte, &userInfo)
	if err != nil {
		return err
	}

	user.CopyValue(userInfo)
	return nil
}

func (user *UserDetailV3) CopyValue(source UserDetailV3) {
	user.ID = source.ID
	user.DisplayName = source.DisplayName
	user.Email = source.Email
	user.Enterprise = source.Enterprise
	user.Type = source.Type
	user.Password = source.Password
	user.Status = source.Status
}
