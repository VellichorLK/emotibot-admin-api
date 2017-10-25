package util

import (
	"io/ioutil"
	"strings"
)

var envs = make(map[string]interface{})

// LoadConfigFromFile will get environment variables from file into envs
// Format in file:
// 	SERVER_ENV_NAME=XXXX -> global env for server
//  MODULE_ENV_NAME=XXXX -> env for specific module
func LoadConfigFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

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
		LogTrace.Printf("Params: %+v\n", params)

		// skip error format
		if len(params) < 2 {
			continue
		}

		key, val := strings.TrimSpace(params[0]), strings.TrimSpace(params[1])
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
		moduleEnv[newKey] = strings.Trim(val, "\"")
	}
	LogTrace.Printf("Load config: %+v\n", envs)

	return nil
}

func GetEnvOf(module string) map[string]string {
	if envMap, ok := envs[module]; ok {
		return envMap.(map[string]string)
	}
	return make(map[string]string)
}
