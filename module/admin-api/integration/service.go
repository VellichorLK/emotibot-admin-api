package integration

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/module/admin-api/integration/internal/datatype"
)

type AnswerObj struct {
	Type    string `json:"type"`
	SubType string `json:"subType"`
	Value   string `json:"value"`
}

const PlatformWorkWeixin = "workweixin"

func (ans AnswerObj) IsSupport() bool {
	return ans.Type == "text"
}

func (ans AnswerObj) ToString() string {
	return ans.Value
}

func GetChatResult(appid, userid, input string) string {
	conf := &QA.QATestInput{}
	conf.UserInput = input
	answer, _, err := QA.DoChatRequestWithOpenAPI(appid, userid, conf)
	if err != nil {
		return fmt.Sprintf("System error: %s", err.Error())
	}
	if answer.Answers == nil || len(answer.Answers) == 0 {
		return "No response"
	}

	ret := []string{}
	for idx := range answer.Answers {
		t := AnswerObj{}
		err := json.Unmarshal([]byte(*answer.Answers[idx]), &t)
		if err != nil {
			ret = append(ret, *answer.Answers[idx])
		} else {
			ret = append(ret, t.ToString())
		}
	}

	if len(ret) == 0 {
		return "No valid response"
	}

	return strings.Join(ret, "\n")
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
