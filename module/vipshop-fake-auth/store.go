package main

var (
	Privileges = map[string]PrivilegeRet{
		"view-log": PrivilegeRet{
			PrivilegeName: "view-log",
			AssetName:     "日志管理",
		},
		"export-log": PrivilegeRet{
			PrivilegeName: "export-log",
			AssetName:     "日志管理",
		},
		"view-analysis": PrivilegeRet{
			PrivilegeName: "view-analysis",
			AssetName:     "统计分析",
		},
		"export-analysis": PrivilegeRet{
			PrivilegeName: "export-analysis",
			AssetName:     "统计分析",
		},
		"view-safety": PrivilegeRet{
			PrivilegeName: "view-safety",
			AssetName:     "安全日志",
		},
		"export-safety": PrivilegeRet{
			PrivilegeName: "export-safety",
			AssetName:     "安全日志",
		},
		"view-qa": PrivilegeRet{
			PrivilegeName: "view-qa",
			AssetName:     "问答管理",
		},
		"export-qa": PrivilegeRet{
			PrivilegeName: "export-qa",
			AssetName:     "问答管理",
		},
		"import-qa": PrivilegeRet{
			PrivilegeName: "import-qa",
			AssetName:     "问答管理",
		},
		"add-qa": PrivilegeRet{
			PrivilegeName: "add-qa",
			AssetName:     "问答管理",
		},
		"modify-qa": PrivilegeRet{
			PrivilegeName: "modify-qa",
			AssetName:     "问答管理",
		},
		"delete-qa": PrivilegeRet{
			PrivilegeName: "delete-qa",
			AssetName:     "问答管理",
		},
		"view-qatest": PrivilegeRet{
			PrivilegeName: "view-qatest",
			AssetName:     "对话测试",
		},
		"view-dict": PrivilegeRet{
			PrivilegeName: "view-dict",
			AssetName:     "词库管理",
		},
		"export-dict": PrivilegeRet{
			PrivilegeName: "export-dict",
			AssetName:     "词库管理",
		},
		"import-dict": PrivilegeRet{
			PrivilegeName: "import-dict",
			AssetName:     "词库管理",
		},
		"view-profile": PrivilegeRet{
			PrivilegeName: "view-profile",
			AssetName:     "形象设置",
		},
		"modify-profile": PrivilegeRet{
			PrivilegeName: "modify-profile",
			AssetName:     "形象设置",
		},
		"view-skill": PrivilegeRet{
			PrivilegeName: "view-skill",
			AssetName:     "技能设置",
		},
		"modify-skill": PrivilegeRet{
			PrivilegeName: "modify-skill",
			AssetName:     "技能设置",
		},
		"view-answer": PrivilegeRet{
			PrivilegeName: "view-answer",
			AssetName:     "话术设置",
		},
		"modify-answer": PrivilegeRet{
			PrivilegeName: "modify-answer",
			AssetName:     "话术设置",
		},
		"view-switch": PrivilegeRet{
			PrivilegeName: "view-switch",
			AssetName:     "开关管理",
		},
		"modify-switch": PrivilegeRet{
			PrivilegeName: "modify-switch",
			AssetName:     "开关管理",
		},
		"add-user": PrivilegeRet{
			PrivilegeName: "add-user",
			AssetName:     "权限管理",
		},
		"modify-user": PrivilegeRet{
			PrivilegeName: "modify-user",
			AssetName:     "权限管理",
		},
		"delete-user": PrivilegeRet{
			PrivilegeName: "delete-user",
			AssetName:     "权限管理",
		},
		"modify-role": PrivilegeRet{
			PrivilegeName: "modify-role",
			AssetName:     "权限管理",
		},
		"modify-taskengine": PrivilegeRet{
			PrivilegeName: "modify-taskengine",
			AssetName:     "任务引擎",
		},
	}
	Roles = map[string]*StoreRole{
		"admin": &StoreRole{
			RoleName:        "admin",
			ApplicationName: "VCA",
			CreateTime:      1509220897,
			LastModifyTime:  1509220897,
			RoleDesc:        "admin",
			RoleState:       1,
			Privileges: []string{
				"view-log", "export-log", "view-analysis", "export-analysis", "view-safety", "export-safety", "view-qa", "export-qa", "import-qa", "add-qa", "modify-qa", "delete-qa", "view-qatest", "view-dict", "export-dict", "import-dict", "view-profile", "modify-profile", "view-skill", "modify-skill", "view-answer", "modify-answer", "view-switch", "modify-switch", "add-user", "modify-user", "delete-user", "modify-role", "modify-taskengine",
			},
		},
		"test": &StoreRole{
			RoleName:        "test",
			ApplicationName: "VCA",
			CreateTime:      1509220897,
			LastModifyTime:  1509220897,
			RoleDesc:        "test",
			RoleState:       1,
			Privileges: []string{
				"view-log",
			},
		},
	}
	Users = map[string]*StoreUser{
		"user1": &StoreUser{
			UserName:       "测试1",
			UserDepartment: "test",
			UserAccountID:  "user1",
			UserCode:       "0",
			Roles:          []string{"admin"},
		},
		"user2": &StoreUser{
			UserName:       "测试2",
			UserDepartment: "test",
			UserAccountID:  "user2",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user3": &StoreUser{
			UserName:       "测试3",
			UserDepartment: "test",
			UserAccountID:  "user3",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user4": &StoreUser{
			UserName:       "测试4",
			UserDepartment: "test",
			UserAccountID:  "user4",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user5": &StoreUser{
			UserName:       "测试5",
			UserDepartment: "test",
			UserAccountID:  "user5",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user6": &StoreUser{
			UserName:       "测试6",
			UserDepartment: "test",
			UserAccountID:  "user6",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user7": &StoreUser{
			UserName:       "测试7",
			UserDepartment: "test",
			UserAccountID:  "user7",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user8": &StoreUser{
			UserName:       "测试8",
			UserDepartment: "test",
			UserAccountID:  "user8",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user9": &StoreUser{
			UserName:       "测试9",
			UserDepartment: "test",
			UserAccountID:  "user9",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user10": &StoreUser{
			UserName:       "测试10",
			UserDepartment: "test",
			UserAccountID:  "user10",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user11": &StoreUser{
			UserName:       "测试11",
			UserDepartment: "test",
			UserAccountID:  "user11",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
		"user1+1": &StoreUser{
			UserName:       "测试1+1",
			UserDepartment: "test",
			UserAccountID:  "user12",
			UserCode:       "0",
			Roles:          []string{"test"},
		},
	}
)

func Contains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func Remove(arr []string, str string) []string {
	idx := -1
	for i, s := range arr {
		if s == str {
			idx = i
			break
		}
	}

	if idx == -1 {
		return arr
	}

	arr[idx] = arr[len(arr)-1]
	return arr[:len(arr)-1]
}
