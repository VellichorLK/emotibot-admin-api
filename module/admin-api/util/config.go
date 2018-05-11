package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var envs = make(map[string]interface{})

const encryptedSettingKey = "DECRYPTION_SERVICE"
const sqlPasswordKey = "MYSQL_PASS"

// LoadConfigFromFile will get environment variables from file into envs
// Format in file:
// 	SERVER_ENV_NAME=XXXX -> global env for server
//  MODULE_ENV_NAME=XXXX -> env for specific module
func LoadConfigFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	useEntryptedPassword := (strings.TrimSpace(os.Getenv(encryptedSettingKey)) != "")

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// skip empty line
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// skip comment
		if line[0] == '#' {
			continue
		}

		params := strings.Split(line, "=")

		// skip error format
		if len(params) < 2 {
			continue
		}

		key, val := strings.TrimSpace(params[0]), strings.TrimSpace(strings.Join(params[1:], "="))
		keyParts := strings.Split(key, "_")

		//skip error format
		if len(keyParts) < 2 {
			continue
		}

		envType := strings.ToLower(keyParts[0])
		newKey := strings.Join(keyParts[1:], "_")
		if _, ok := envs[envType]; !ok {
			envs[envType] = make(map[string]string)
		}

		moduleEnv := envs[envType].(map[string]string)
		setValue := strings.Trim(val, "\"")
		if useEntryptedPassword && isSQLPasswordKey(newKey) {
			newSetValue, err := DesDecrypt(setValue, []byte(DesEncryptKey))
			if err == nil {
				LogTrace.Printf("Decrypt password %s => %s\n", setValue, newSetValue)
				setValue = newSetValue
			} else {
				LogError.Printf("Decrypt password error %s: %s\n", setValue, err.Error())
			}
		}
		moduleEnv[newKey] = setValue
	}

	envsStr, err := json.MarshalIndent(envs, "", "  ")
	LogInfo.Printf("Load config: %s\n", envsStr)

	return nil
}

func GetEnvOf(module string) map[string]string {
	if envMap, ok := envs[module]; ok {
		return envMap.(map[string]string)
	}
	return make(map[string]string)
}

func getGlobalEnv(key string) string {
	envs := GetEnvOf("server")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func isSQLPasswordKey(key string) (ret bool) {
	words := strings.Split(key, "_")
	l := len(words)
	if l < 2 {
		return false
	}

	ret = (sqlPasswordKey == fmt.Sprintf("%s_%s", words[l-2], words[l-1]))
	return
}
