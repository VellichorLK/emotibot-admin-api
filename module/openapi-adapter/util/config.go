package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

var envs = make(map[string]string)
var envPrefix = "OPENAPI_ADAPTER"

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
	for _, line := range lines {
		// Skip empty line
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Skip comment
		if line[0] == '#' {
			continue
		}

		params := strings.Split(line, "=")

		// Skip error format
		if len(params) < 2 {
			continue
		}

		// Key: remove 'OPENAPI_ADAPTER' prefix
		key := strings.ToUpper(strings.TrimSpace(params[0]))
		keyParts := strings.Split(key, "_")

		// Skip error format
		if len(keyParts) < 2 {
			continue
		}

		if strings.Join(keyParts[0:2], "_") != envPrefix {
			continue
		}

		val := params[1]
		envs[strings.Join(keyParts[2:], "_")] = val
	}

	envsStr, _ := json.MarshalIndent(envs, "", "  ")
	logger.Info.Printf("Load config: %s\n", envsStr)

	return nil
}

func GetEnv(key string) (string, bool) {
	val, ok := envs[key]
	return val, ok
}

func GetIntEnv(key string) (int, error) {
	valStr, ok := GetEnv(key)
	if !ok {
		errMsg := fmt.Sprintf("No %s env variable\n", key)
		return 0, errors.New(errMsg)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, err
	}

	return val, nil
}
