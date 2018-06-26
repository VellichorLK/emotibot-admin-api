package Task

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"runtime"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func EnableAllScenario(appid string) {
	setAllScenarioStatus(appid, true)
}

func DisableAllScenario(appid string) {
	setAllScenarioStatus(appid, false)
}

func EnableScenario(appid string, scenarioID string) {
	setScenarioStatus(appid, scenarioID, true)
}

func DisableScenario(appid string, scenarioID string) {
	setScenarioStatus(appid, scenarioID, false)
}

func ReadUploadJSON(file multipart.File) (string, error) {
	buf := bytes.NewBuffer(nil)
	size, err := io.Copy(buf, file)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", fmt.Errorf("Size of file is zero")
	}
	return buf.String(), nil
}

func ImportScenarios(appid string, newID bool, datas []interface{}) {
	for _, data := range datas {
		ImportScenario(appid, newID, data)
	}
}

func ImportScenario(appid string, newID bool, data interface{}) {
	// util.LogInfo.Printf("Import scenario, use new ID: %t \ndata: %#v\n", newID, data)
	if !newID {
		updateScenario(appid, data)
	} else {
		ret, err := addScenario(appid, data)
		if err != nil {
			fmt.Printf("Add scenario content: %s, error: %s\n", ret, err.Error())
		}
	}
}

func parseJSONData(data interface{}) (string, string, interface{}, interface{}) {
	var content map[string]interface{}
	var layout interface{}
	taskEngineData, ok := data.(*map[string]interface{})
	if ok {
		content = (*taskEngineData)["taskScenario"].(map[string]interface{})
		layout = (*taskEngineData)["taskLayouts"]
	} else {
		taskEngineData := data.(map[string]interface{})
		content = taskEngineData["taskScenario"].(map[string]interface{})
		layout = taskEngineData["taskLayouts"]
	}
	metadata := content["metadata"].(map[string]interface{})
	id := metadata["scenario_id"].(string)
	name := metadata["scenario_name"].(string)
	return id, name, content, layout
}

func updateScenario(appid string, data interface{}) (string, error) {
	id, _, content, layout := parseJSONData(data)
	existed, err := checkScenarioExist(appid, id)
	if err != nil {
		return "", err
	}
	if !existed {
		return addScenario(appid, data)
	}

	postData := map[string]interface{}{
		"userid":  appid,
		"content": content,
		"layout":  layout,
	}
	url := fmt.Sprintf("%s/%s/%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id)
	ret, err := util.HTTPPutForm(url, postData, 0)
	return ret, err
}
func addScenario(appid string, data interface{}) (retStr string, err error) {
	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		if r, ok := recover().(error); ok {
			_, file, line, _ := runtime.Caller(1)
			fmt.Printf("Recovered in %s:%d - %s\n", file, line, r)
			err = errors.New("Handle add scenario fail")
		}
	}()
	_, name, content, layout := parseJSONData(data)

	postData := map[string]string{
		"userid":       appid,
		"scenarioname": name,
	}
	url := fmt.Sprintf("%s/%s", getEnvironment("SERVER_URL"), taskScenarioEntry)
	retData, err := util.HTTPPostForm(url, postData, 0)
	if err != nil {
		util.LogTrace.Printf("Err: %s", err.Error())
		return "", err
	}
	retObj := map[string]interface{}{}
	err = json.Unmarshal([]byte(retData), &retObj)
	if err != nil {
		util.LogTrace.Printf("Err: %s", err.Error())
		return "", err
	}
	id := retObj["scenarioID"].(string)
	contentObj := content.(map[string]interface{})
	metadata := contentObj["metadata"].(map[string]interface{})
	util.LogInfo.Printf("metadata: %#v\n", metadata)
	metadata["scenario_id"] = id

	updateData := map[string]interface{}{
		"userid":  appid,
		"content": content,
		"layout":  layout,
	}
	url = fmt.Sprintf("%s/%s/%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id)
	util.LogTrace.Printf("Call %s to addScenario\n", url)
	retStr, err = util.HTTPPutForm(url, updateData, 0)
	return
}

func checkScenarioExist(appid string, id string) (bool, error) {
	values := url.Values{
		"userid": []string{appid},
	}
	url := fmt.Sprintf("%s/%s/%s?%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id,
		values.Encode())
	content, err := util.HTTPGetSimple(url)
	if err != nil {
		util.LogInfo.Printf("Check scenario %s existed fail: %s", id, err.Error())
		return false, err
	}

	ret := map[string]*interface{}{}
	err = json.Unmarshal([]byte(content), &ret)
	if err != nil {
		util.LogInfo.Printf("Check scenario %s existed fail: %s", id, err.Error())
		return false, err
	}

	if ret["result"] == nil {
		util.LogTrace.Printf("Check scenario: %s not exist\n", id)
		return false, nil
	}
	util.LogTrace.Printf("Check scenario: %s exist\n", id)
	return true, nil
}

func GetMapTableList(appid, userID string) ([]string, int, error) {
	list, err := getMapTableList(appid, userID)
	if err == sql.ErrNoRows {
		return []string{}, ApiError.SUCCESS, nil
	} else if err != nil {
		return []string{}, ApiError.DB_ERROR, err
	}
	return list, ApiError.SUCCESS, nil
}

func GetMapTableContent(appid, userID, tableName string) (string, int, error) {
	content, err := getMapTableContent(appid, userID, tableName)
	if err == sql.ErrNoRows {
		return "", ApiError.NOT_FOUND_ERROR, err
	} else if err != nil {
		return "", ApiError.DB_ERROR, err
	}
	return content, ApiError.SUCCESS, nil
}

func SaveMappingTable(userID, appid, fileName, content string) (int, error) {
	err := saveMappingTable(userID, appid, fileName, content)
	if err != nil {
		return ApiError.DB_ERROR, nil
	}
	return ApiError.SUCCESS, nil
}

func ParseUploadMappingTable(buf []byte) ([]*MapTuple, error) {
	// REMOVE utf-8 BOM if existed
	buf = bytes.Trim(buf, "\xef\xbb\xbf")

	buf = bytes.Replace(buf, []byte("\r\n"), []byte("\n"), -1)
	buf = bytes.Replace(buf, []byte("\r"), []byte("\n"), -1)

	lines := strings.Split(string(buf), "\n")

	ret := []*MapTuple{}
	for idx, line := range lines {
		params := strings.Split(line, ",")
		if len(params) != 2 {
			util.LogTrace.Printf("Parse csv in line %d format error", idx+1)
			continue
		}

		temp := MapTuple{
			Key:   params[0],
			Value: params[1],
		}
		ret = append(ret, &temp)
	}
	return ret, nil
}

func DeleteMappingTable(appid, userID, tableName string) error {
	err := deleteMappingTable(appid, userID, tableName)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
