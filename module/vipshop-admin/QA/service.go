package QA

import (
	"encoding/json"
	"errors"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func DoChatRequest(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	openAPIURL := getTestURL()
	if len(openAPIURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	// Prepare for openAPI input
	input := make(map[string]string)
	input["userid"] = user
	input["appid"] = appid
	input["text"] = inputData.UserInput
	input["cmd"] = "chat"

	customInfo := map[string]string{
		"qtype":    "debug",
		"top":      "2",
		"platform": inputData.Platform,
		"brand":    inputData.Brand,
		"sex":      inputData.Gender,
		"age":      inputData.Age,
		"hobbies":  inputData.Hobbies,
	}
	customInfoStr, _ := json.Marshal(customInfo)

	customReturn := map[string]string{
		"nodeId":     "source",
		"Module":     "module",
		"SubModule":  "customReturn",
		"Similarity": "score",
		"cu_word":    "CU/wordPos",
	}
	customReturnStr, _ := json.Marshal(customReturn)

	input["customInfo"] = string(customInfoStr)
	input["customReturn"] = string(customReturnStr)

	util.LogTrace.Printf("CustomInfo: %s\n", customInfoStr)
	util.LogTrace.Printf("CustomReturn: %s\n", customReturnStr)

	response, err := util.HTTPPostForm(openAPIURL, input, 10)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, nil
	}
	util.LogTrace.Printf("Raw response from OpenAPI: %s", response)

	openAPIRet := OpenAPIResponse{}
	err = json.Unmarshal([]byte(response), &openAPIRet)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}

	ret := RetData{}
	if openAPIRet.Emotion != nil && len(openAPIRet.Emotion) > 0 {
		ret.Emotion = openAPIRet.Emotion[0].Value
	}
	if openAPIRet.Intent != nil && len(openAPIRet.Intent) > 0 {
		ret.Intent = openAPIRet.Intent[0].Value
	}
	if openAPIRet.Data != nil && len(openAPIRet.Data) > 0 {
		ret.Answers = []*string{}
		for idx := range openAPIRet.Data {
			ret.Answers = append(ret.Answers, &openAPIRet.Data[idx].Value)

		}
	} else {
		return nil, ApiError.QA_TEST_FORMAT_ERROR, errors.New("Answer column is empty")
	}
	ret.OpenAPIReturn = openAPIRet.Return
	ret.SimilarQuestion = getSimilarFromCustomReturn(&openAPIRet)
	if ret.SimilarQuestion == nil {
		ret.SimilarQuestion = []*QuestionInfo{}
	}
	ret.Tokens = getTokensFromCustomReturn(&openAPIRet)
	if ret.Tokens == nil {
		ret.Tokens = []*string{}
	}

	return &ret, ApiError.SUCCESS, nil
}

func getSimilarFromCustomReturn(res *OpenAPIResponse) []*QuestionInfo {
	defer func() {
		if r := recover(); r != nil {
			util.LogInfo.Println("Parse from openapi's customReturn error: ", r)
		}
	}()

	ret := []*QuestionInfo{}

	responseInfo := res.CustomReturn["response_other_info"].(map[string]interface{})
	relatedQuestion := responseInfo["relatedQ"].([]interface{})
	for _, val := range relatedQuestion {
		temp := val.(map[string]interface{})
		similarQ := QuestionInfo{}
		if temp["stdQ"] != nil {
			similarQ.StandardQuestion = temp["stdQ"].(string)
		}
		if temp["userQ"] != nil {
			similarQ.UserQuestion = temp["userQ"].(string)
		}
		if temp["score"] != nil {
			similarQ.SimilarityScore = temp["score"].(float64)
		}
		ret = append(ret, &similarQ)
	}
	return ret
}

func getTokensFromCustomReturn(res *OpenAPIResponse) []*string {
	defer func() {
		if r := recover(); r != nil {
			util.LogInfo.Println("Parse from openapi's customReturn error: ", r)
		}
	}()

	ret := []*string{}

	responseInfo := res.CustomReturn["response_other_info"].(map[string]interface{})
	tokens := responseInfo["token"].([]interface{})
	for _, val := range tokens {
		temp := val.(string)
		ret = append(ret, &temp)
	}
	return ret
}
