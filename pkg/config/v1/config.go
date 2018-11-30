package config

import (
	"io/ioutil"
	"encoding/json"
	"os"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

var envs = make(map[string]map[string]string)

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

// storeEnvs will store input envs with  format like MODNAME_XXX_YYY_ZZZZ=kkkk
func storeEnvs(origEnvs map[string]string) error {
	for key, val := range origEnvs {
		keyParts := strings.Split(key, "_")

		// skip error format
		// env key format must be MODNAME_XXX_YYYY...
		if len(keyParts) < 2 {
			continue
		}

		envType := strings.ToLower(keyParts[0])
		newKey := strings.Join(keyParts[1:], "_")
		if _, ok := envs[envType]; !ok {
			envs[envType] = make(map[string]string)
		}

		moduleEnv := envs[envType]
		setValue := strings.Trim(val, "\"")
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