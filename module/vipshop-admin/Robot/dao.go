package Robot

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func getFunctionList(appid string) (map[string]*FunctionInfo, error) {
	filePath := util.GetFunctionSettingPath(appid)
	ret := make(map[string]*FunctionInfo)

	// If file not exist, return empty map
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		util.LogInfo.Printf("File of function setting not existed for appid = [%s]", filePath)
		return ret, nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// each line is function_name = on/off pair
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		pair := strings.Split(line, "=")
		if len(pair) == 2 {
			append := FunctionInfo{}
			function := strings.Trim(pair[0], " \"")
			value := strings.Trim(pair[1], " \"")
			if value == "true" || value == "on" || value == "1" {
				append.Status = true
			} else {
				append.Status = false
			}
			ret[function] = &append
		}
	}

	return ret, nil
}

func updateFunctionList(appid string, infos map[string]*FunctionInfo) error {
	filePath := util.GetFunctionSettingTmpPath(appid)

	lines := []string{}
	for key, info := range infos {
		lines = append(lines, fmt.Sprintf("%s = %t\n", key, info.Status))
	}
	content := strings.Join(lines, "")
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	util.LogTrace.Printf("Upload function properties to %s", content)

	now := time.Now()
	nowStr := now.Format("2000-01-01 00:00:00")
	ioutil.WriteFile(filePath, []byte(nowStr), 0644)

	return err
}
