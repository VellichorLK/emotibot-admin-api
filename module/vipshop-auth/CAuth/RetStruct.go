package CAuth

import "emotibot.com/emotigo/module/vipshop-admin/ApiError"

type RetObj struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

type Privilege struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	CmdList string `json:"cmdlist"`
}

type Role struct {
	RoleID    string `json:"role_id"`
	RoleName  string `json:"role_name"`
	Privilege string `json:"privilege"`
}

type UserProp struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
	UserType int    `json:"user_type"`
	RoleId   string `json:"role_id"`
}

func GenRetObj(status int, obj interface{}) RetObj {
	return RetObj{
		Status:  status,
		Message: ApiError.GetErrorMsg(status),
		Result:  obj,
	}
}

func GenRole(id string, name string, privilege string) *Role {
	return &Role{
		RoleID:    id,
		RoleName:  name,
		Privilege: privilege,
	}
}

func GenPrivilege(id int, name string, cmdList string) *Privilege {
	return &Privilege{
		ID:      id,
		Name:    name,
		CmdList: cmdList,
	}
}
