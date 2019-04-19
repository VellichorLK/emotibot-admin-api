package integration

import (
	"encoding/json"
	"fmt"

	"emotibot.com/emotigo/pkg/misc/adminerrors"

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
