package Service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo        util.ModuleInfo
	serviceNLUKey     = "NLU"
	serviceSolrETLKey = "SOLRETL"
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName:  "service",
		EntryPoints: []util.EntryPoint{},
	}
}

func GetNLUResult(appid string, sentence string) (*NLUResult, error) {
	url := strings.TrimSpace(getEnvironment(serviceNLUKey))
	if url == "" {
		return nil, errors.New("NLU Service not set")
	}
	param := map[string]string{
		"f":     "segment,sentenceType,keyword",
		"appid": appid,
		"q":     sentence,
	}
	body, err := util.HTTPGet(url, param, 30)
	if err != nil {
		return nil, err
	}

	nluResult := []NLUResult{}
	err = json.Unmarshal([]byte(body), &nluResult)
	if err != nil {
		return nil, err
	}
	if len(nluResult) < 1 {
		return nil, errors.New("No result")
	}
	return &nluResult[0], nil
}

func BatchGetNLUResults(appid string, sentences []string) (map[string]*NLUResult, error) {
	sentencePerReq := 20
	if len(sentences) < sentencePerReq {
		return GetNLUResults(appid, sentences)
	}

	dataChan := make(chan []string)
	resultsChan := make(chan (*map[string]*NLUResult))
	allResult := map[string]*NLUResult{}
	defer func() {
		close(dataChan)
		close(resultsChan)
	}()

	maxWorker := 5
	for idx := 0; idx < maxWorker; idx++ {
		go func(workNo int) {
			for {
				sentencesGroup, ok := <-dataChan
				// if channel close, just exit
				if !ok {
					return
				}
				util.LogTrace.Printf("Worker %d receive %d sentences\n", workNo, len(sentencesGroup))
				ret, err := GetNLUResults(appid, sentencesGroup)
				if err != nil {
					util.LogError.Println("Get NLU Result error:", err.Error())
					ret = map[string]*NLUResult{}
				}
				util.LogTrace.Printf("Worker %d finish query NLU, get %#v\n", workNo, ret)
				resultsChan <- &ret
			}
		}(idx)
	}

	packetNum := 0
	for idx := 0; idx < len(sentences); idx += sentencePerReq {
		ending := idx + sentencePerReq
		if ending > len(sentences) {
			ending = len(sentences)
		}
		util.LogTrace.Printf("Send sentence %d-%d into channel\n", idx, ending)
		dataChan <- sentences[idx:ending]
		packetNum++
	}
	util.LogTrace.Printf("Send %d packets into channel\n", packetNum)

	for packetNum > 0 {
		groupResult := <-resultsChan
		packetNum--
		util.LogTrace.Printf("Master get %#v\n", groupResult)
		for k, v := range *groupResult {
			allResult[k] = v
		}
	}

	return allResult, nil
}

func GetNLUResults(appid string, sentences []string) (map[string]*NLUResult, error) {
	url := strings.TrimSpace(getEnvironment(serviceNLUKey))
	if url == "" {
		return nil, errors.New("NLU Service not set")
	}
	param := map[string]interface{}{
		"flags":   "segment,sentenceType,keyword",
		"appid":   appid,
		"queries": sentences,
	}
	body, err := util.HTTPPostJSON(url, param, 30)
	// body, err := util.HTTPGet(url, param, 30)
	if err != nil {
		return nil, err
	}

	nluResult := []*NLUResult{}
	util.LogTrace.Println("NLU Response:", body)
	err = json.Unmarshal([]byte(body), &nluResult)
	if err != nil {
		return nil, err
	}
	if len(nluResult) < 1 {
		return nil, errors.New("No result")
	}

	ret := map[string]*NLUResult{}
	for idx, result := range nluResult {
		ret[result.Sentence] = nluResult[idx]
	}
	return ret, nil
}

func IncrementAddSolr(content []byte) (string, error) {
	url := getSolrIncrementURL()
	if url == "" {
		return "", errors.New("Solr-etl Service not set")
	}

	reader := bytes.NewReader(content)
	status, body, err := util.HTTPPostFileWithStatus(url, reader, "robot_manual_tagging.json", "file", 30)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return body, fmt.Errorf("Status not 200, is %d", status)
	}
	return body, nil
}

func DeleteInSolr(deleteSolrIDs []string) (string, error) {
	url := getSolrDeleteURL()
	params := map[string]string{
		"ids": strings.Join(deleteSolrIDs, ","),
	}
	return util.HTTPGet(url, params, 30)
}

func getSolrIncrementURL() string {
	url := strings.TrimSpace(getEnvironment(serviceSolrETLKey))
	return fmt.Sprintf("%s/editorialincre", url)
}
func getSolrDeleteURL() string {
	url := strings.TrimSpace(getEnvironment(serviceSolrETLKey))
	return fmt.Sprintf("%s/deletebyids", url)
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}
