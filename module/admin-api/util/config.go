package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

var envs = make(map[string]map[string]string)

const encryptedSettingKey = "DECRYPTION_SERVICE"
const sqlPasswordKey = "MYSQL_PASS"
const envPrefix = "admin"

// LoadConfigFromFile will get environment variables from file into envs
func LoadConfigFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	return parseEnvLines(lines)
}

// LoadConfigFromOSEnv will get environment variables from os environment
func LoadConfigFromOSEnv() error {
	return parseEnvLines(os.Environ())
}

// parseEnvLines will transform env lines to saved environment variables
func parseEnvLines(lines []string) error {
	fileEnvs := map[string]string{}
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
		fileEnvs[key] = val
	}

	return storeEnvs(fileEnvs)
}

// storeEnvs will store input envs with belowing format
// Format in environment variable:
// 	ADMIN_SERVER_ENV_NAME=XXXX -> global env for server
//  ADMIN_MODULE_ENV_NAME=XXXX -> env for specific module
func storeEnvs(origEnvs map[string]string) error {
	useEntryptedPassword := (strings.TrimSpace(os.Getenv(encryptedSettingKey)) != "")
	for key, val := range origEnvs {
		keyParts := strings.Split(key, "_")

		// skip error format
		// env key format must be ADMIN_MODNAME_XXX_...
		if len(keyParts) < 3 {
			continue
		}

		if strings.ToLower(keyParts[0]) != envPrefix {
			continue
		}

		envType := strings.ToLower(keyParts[1])
		newKey := strings.Join(keyParts[2:], "_")
		if _, ok := envs[envType]; !ok {
			envs[envType] = make(map[string]string)
		}

		moduleEnv := envs[envType]
		setValue := strings.Trim(val, "\"")
		if useEntryptedPassword && isSQLPasswordKey(newKey) {
			newSetValue, err := DesDecrypt(setValue, []byte(DesEncryptKey))
			if err == nil {
				logger.Trace.Printf("Decrypt password %s => %s\n", setValue, newSetValue)
				setValue = newSetValue
			} else {
				logger.Error.Printf("Decrypt password error %s: %s\n", setValue, err.Error())
			}
		}
		moduleEnv[newKey] = setValue
	}

	envsStr, _ := json.MarshalIndent(envs, "", "  ")
	logger.Info.Printf("Load config: %s\n", envsStr)
	return nil
}

func GetEnvOf(module string) map[string]string {
	if envMap, ok := envs[module]; ok {
		return envMap
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
