package Task

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

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
			fmt.Println("Recovered in f: ", r)
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
	if err == nil {
		return "", err
	}
	retObj := map[string]interface{}{}
	err = json.Unmarshal([]byte(retData), &retObj)
	if err == nil {
		return "", err
	}
	id := retObj["scenarioID"].(string)
	contentObj := content.(map[string]interface{})
	util.LogInfo.Printf("content: %#v\n", contentObj)
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
	url := fmt.Sprintf("%s/%s/%s?userid=%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id,
		appid)
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
