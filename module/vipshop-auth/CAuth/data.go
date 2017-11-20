package CAuth

import "emotibot.com/emotigo/module/vipshop-admin/util"

const cAuthServerKey = "SERVER_ADDR"
const requestHashKey = "APP_KEY"
const requestRequesterKey = "REQUESTER"
const cAuthPrefix = "PREFIX"

var (
	// PrivilegesMap is a map, key is module from cauth return, value is the same in origin system
	PrivilegesMap = map[string]*Privilege{
		"log":        GenPrivilege(1, "statistic-daily", "view,export"),
		"analysis":   GenPrivilege(2, "statistic-analysis", "view,export"),
		"safety":     GenPrivilege(3, "statistic-audit", "view,export"),
		"qa":         GenPrivilege(4, "qalist", "view,modify,add,delete,export,import"),
		"qatest":     GenPrivilege(5, "qatest", "view"),
		"dict":       GenPrivilege(6, "dictionary", "view,import,export"),
		"profile":    GenPrivilege(7, "robot-profile", "view,modify"),
		"skill":      GenPrivilege(8, "robot-skill", "view,modify"),
		"answer":     GenPrivilege(9, "robot-chat", "view,edit"),
		"switch":     GenPrivilege(10, "switch-manage", "view,edit"),
		"user":       GenPrivilege(11, "user-manage", "modify"),
		"role":       GenPrivilege(12, "role-manage", "modify"),
		"taskengine": GenPrivilege(13, "task-engine", "view,edit"),
	}
)

func getEnvironment(key string) string {
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func getCAuthServer() string {
	return getEnvironment(cAuthServerKey)
}

func getCAuthAppKey() string {
	return getEnvironment(requestHashKey)
}

func getCAuthRequester() string {
	return getEnvironment(requestRequesterKey)
}

func getCAuthPrefix() string {
	return getEnvironment(cAuthPrefix)
}
