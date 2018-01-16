package util

import (
	"strings"
)

//GetModuleEnvironments get the modules' enviroment and put it into map
func GetModuleEnvironments(module string) map[string]string {
	return GetEnvOf(strings.ToLower(module))
}

//GetEnviroment get enviroment variable from given parameter env
func GetEnviroment(envs map[string]string, key string) string {
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}
