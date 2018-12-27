package integration

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/module/admin-api/integration/internal/datatype"
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

func GetPlatformConfig(appid, platform string) (map[string]string, error) {
	return getPlatformConfig(appid, platform)
}

func GetQuestion(r *http.Request, platform string, config map[string]string) (string, error) {
	switch platform {
	case PlatformWorkWeixin:
		return getWorkWeixinQuestion(r, config)
	}
	return "", errors.New("Unsupported platform")
}
func getWorkWeixinQuestion(r *http.Request, config map[string]string) (string, error) {
	platformConfig := datatype.WorkWeiXinConfig{}
	platformConfig.Load(config)
	if !platformConfig.IsValid() {
		return "", errors.New("Config invalid")
	}

	content, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return "", err
	}

	request := datatype.WorkWeixinRequest{}
	err = xml.Unmarshal(content, &request)
	if err != nil {
		return "", err
	}

	// TODO
	return "Hardcode input", nil
}

func Sendback(w http.ResponseWriter, platform, answer string, config map[string]string) error {
	switch platform {
	case PlatformWorkWeixin:
		return SendbackWorkWeixin(w, answer, config)
	}
	return errors.New("Unsupported platform")
}

func SendbackWorkWeixin(w http.ResponseWriter, answer string, config map[string]string) error {
	return nil
}
