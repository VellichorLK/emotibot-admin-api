package main

// =====================================
// Struct used in getPrivilegesByRole
// =====================================

// RolePrivilegesParam used in post body
type RolePrivilegesParam struct {
	RoleName        string `json:"roleName"`
	ApplicationName string `json:"applicationName"`
	AppKey          string `json:"appKey"`
}

// PrivilegeRet is element in PrivilegesRet
type PrivilegeRet struct {
	PrivilegeName string `json:"privilegeName"`
	AssetName     string `json:"assetName"`
}

// PrivilegesRet is struct returned by getPrivilegesByRole
type PrivilegesRet struct {
	Data []*PrivilegeRet `json:"data"`
}

// =====================================
// Struct used in getRolesByUsers
// =====================================

// UserRolesParam used in post body
type UserRolesParam struct {
	UserAccounts    []string `json:"userAccounts"`
	ApplicationName string   `json:"applicationName"`
	AppKey          string   `json:"appKey"`
}

// SimpleRoleRet is element in AllRolesRet
type SimpleRoleRet struct {
	RoleName       string `json:"roleName"`
	CreateTime     int64  `json:"createTime"`
	LastModifyTime int64  `json:"lastModifyTime"`
	RoleDesc       string `json:"roleDesc"`
}

// UserRolesRet is struct returned by getAllRolesByAppName
type UserRolesRet struct {
	Data map[string][]*SimpleRoleRet `json:"data"`
}

// =====================================
// Struct used in getAllRolesByAppName
// =====================================

// RolesParam used in post body
type RolesParam struct {
	ApplicationName string `json:"applicationName"`
	AppKey          string `json:"appKey"`
}

// RoleRet is element in AllRolesRet
type RoleRet struct {
	RoleName        string `json:"roleName"`
	ApplicationName string `json:"applicationName"`
	CreateTime      int64  `json:"createTime"`
	LastModifyTime  int64  `json:"lastModifyTime"`
	RoleDesc        string `json:"roleDesc"`
	RoleState       int    `json:"roleState"`
}

// AllRolesRet is struct returned by getAllRolesByAppName
type AllRolesRet struct {
	Data []*RoleRet `json:"data"`
}

// =====================================
// Struct used in getUsesByRole
// =====================================

// RoleUsersParam used in post body
type RoleUsersParam struct {
	RoleName        string `json:"roleName"`
	ApplicationName string `json:"applicationName"`
	AppKey          string `json:"appKey"`
}

// UserRet is element in UsersRet
type UserRet struct {
	UserName       string `json:"userName"`
	UserDepartment string `json:"userDepartment"`
	UserAccountID  string `json:"userAcountId"`
	UserCode       int64  `json:"userCode"`
}

// UsersRet is struct returned by getUsesByRole
type UsersRet struct {
	Data []*UserRet `json:"data"`
}

// =====================================
// Struct used in addRolePrivilege/delRolePrivilege
// =====================================

// RolePrivilegeInput used in addRolePrivilege/delRolePrivilege
type RolePrivilegeInput struct {
	RoleName        string `json:"roleName"`
	PrivilegeName   string `json:"privilegeName"`
	ApplicationName string `json:"applicationName"`
	Requestor       string `json:"requestor"`
	AppKey          string `json:"appKey"`
}

// =====================================
// Struct used in addUserRole/delUserRole
// =====================================

// UserRoleInput used in addUserRole/delUserRole
type UserRoleInput struct {
	RoleName        string `json:"roleName"`
	UserAccount     string `json:"userAccount"`
	UserCode        string `json:"userCode"`
	ApplicationName string `json:"applicationName"`
	Requestor       string `json:"requestor"`
	AppKey          string `json:"appKey"`
}

// =====================================
// Struct used in addRole/delRole
// =====================================

// UserRoleInput used in addUserRole/delUserRole
type RoleInput struct {
	RoleName        string `json:"roleName"`
	RoleDesc        string `json:"roleDesc"`
	ApplicationName string `json:"applicationName"`
	Requestor       string `json:"requestor"`
	AppKey          string `json:"appKey"`
}

// =====================================
// Struct used in return struct
// =====================================

// ErrorStatus is return struct when status is not 200
type ErrorStatus struct {
	Error ReturnStatus `json:"error"`
}

// SuccessStatus is return struct when status is 200
type SuccessStatus struct {
	Data ReturnStatus `json:"data"`
}

// ReturnStatus is general return status after
type ReturnStatus struct {
	ResponseCode int    `json:"responseCode"`
	Message      string `json:"message"`
}

// Used in store
type StoreRole struct {
	RoleName        string
	ApplicationName string
	CreateTime      int64
	LastModifyTime  int64
	RoleDesc        string
	RoleState       int
	Privileges      []string
}

type StoreUser struct {
	UserName       string
	UserDepartment string
	UserAccountID  string
	UserCode       int64
	Roles          []string
}
