package CAS

import "emotibot.com/emotigo/module/vipshop-admin/util"

const CASServer = "SERVER_ADDR"
const CASAppid = "APPID"

func getEnvironment(key string) string {
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func getCASServer() string {
	return getEnvironment(CASServer)
}

func getCASAppid() string {
	return getEnvironment(CASAppid)
}
