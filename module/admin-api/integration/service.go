package integration

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/misc/adminerrors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/pkg/logger"
)

const PlatformWorkWeixin = "workweixin"

func genPureTextNode(input string) *QA.BFOPOpenapiAnswer {
	return &QA.BFOPOpenapiAnswer{
		Type:    "text",
		SubType: "text",
		Value:   input,
		Data:    []string{},
	}
}

func GetChatResult(appid, userid, input string) []*QA.BFOPOpenapiAnswer {
	conf := &QA.QATestInput{}
	conf.UserInput = input
	answer, _, err := QA.DoChatRequestWithBFOPOpenAPI(appid, userid, conf)
	if err != nil {
		return []*QA.BFOPOpenapiAnswer{
			genPureTextNode(fmt.Sprintf("System error: %s", err.Error())),
		}
	}
	if answer.Answers == nil || len(answer.Answers) == 0 {
		return []*QA.BFOPOpenapiAnswer{
			genPureTextNode("No response"),
		}
	}

	ret := []*QA.BFOPOpenapiAnswer{}
	for idx := range answer.Answers {
		t := QA.BFOPOpenapiAnswer{}
		err := json.Unmarshal([]byte(*answer.Answers[idx]), &t)
		if err != nil {
			ret = append(ret, &QA.BFOPOpenapiAnswer{
				Type:    "text",
				SubType: "text",
				Value:   *answer.Answers[idx],
				Data:    []string{},
			})
			logger.Trace.Println("Parse json fail:", err.Error())
		} else {
			logger.Trace.Println("Append", t.ToString())
			ret = append(ret, &t)
		}
	}

	if len(ret) == 0 {
		return []*QA.BFOPOpenapiAnswer{
			genPureTextNode("No valid response"),
		}
	}

	return ret
}

func GetPlatformConfig(appid, platform string) (map[string]string, adminerrors.AdminError) {
	configs, err := getPlatformConfig(appid, platform)
	if err != nil {
		return nil, adminerrors.New(adminerrors.ErrnoDBError, err.Error())
	}
	return configs, nil
}

func SetPlatformConfig(w http.ResponseWriter, r *http.Request) ([]map[string]interface{}, adminerrors.AdminError) {
	params := make(map[string]interface{})

	rawJson, _ := ioutil.ReadAll(r.Body)
	logger.Info.Println(rawJson)
	json.Unmarshal(rawJson, &params)

	params["appid"] = requestheader.GetAppID(r)
	params["platform"] = util.GetMuxVar(r, "platform")

	configs, err := setPlatformConfig(params)
	if err != nil {
		return nil, adminerrors.New(adminerrors.ErrnoDBError, err.Error())
	}
	return configs, nil
}

func DeletePlatformConfig(w http.ResponseWriter, r *http.Request) ([]map[string]interface{}, adminerrors.AdminError) {
	params := make(map[string]interface{})
	params["appid"] = requestheader.GetAppID(r)
	params["platform"] = util.GetMuxVar(r, "platform")

	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	params["url"] = scheme + r.Host + "/api/v1/integration/chat/" + params["platform"].(string) + "/" + params["appid"].(string)

	configs, err := deletePlatformConfig(params)
	if err != nil {
		return nil, adminerrors.New(adminerrors.ErrnoDBError, err.Error())
	}
	return configs, nil
}
