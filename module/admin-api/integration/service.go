package integration

import (
	"encoding/json"
	"fmt"

	"emotibot.com/emotigo/pkg/misc/adminerrors"

	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/pkg/logger"
)

const PlatformWorkWeixin = "微信"
const PlatformLine = "line"

func genPureTextNode(input string) *QA.BFOPOpenapiAnswer {
	return &QA.BFOPOpenapiAnswer{
		Type:    "text",
		SubType: "text",
		Value:   input,
		Data:    nil,
	}
}

func GetChatResult(appid, userid, input string, platform string) []*QA.BFOPOpenapiAnswer {
	conf := &QA.QATestInput{}
	conf.UserInput = input
	conf.Platform = platform
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
				Data:    nil,
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

func SetPlatformConfig(appid, platform string, values map[string]string) (map[string]string, adminerrors.AdminError) {
	configs, err := setPlatformConfig(appid, platform, values)
	if err != nil {
		return nil, adminerrors.New(adminerrors.ErrnoDBError, err.Error())
	}
	return configs, nil
}

func DeletePlatformConfig(appid, platform string) adminerrors.AdminError {
	err := deletePlatformConfig(appid, platform)
	if err != nil {
		return adminerrors.New(adminerrors.ErrnoDBError, err.Error())
	}
	return nil
}
