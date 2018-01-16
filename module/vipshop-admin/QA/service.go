package QA

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func DoChatRequestWithDC(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	openAPIURL := getTestURL()
	if len(openAPIURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	// Prepare for DC input
	input := make(map[string]interface{})
	input["UniqueID"] = "test"
	input["Text1"] = inputData.UserInput
	input["robot"] = appid
	input["UserID"] = user

	customInfo := map[string]string{
		"qtype":    "debug",
		"top":      "2",
		"platform": inputData.Platform,
		"brand":    inputData.Brand,
		"sex":      inputData.Gender,
		"age":      inputData.Age,
		"hobbies":  inputData.Hobbies,
	}

	customReturn := map[string]string{
		"nodeId":     "source",
		"Module":     "module",
		"SubModule":  "customReturn",
		"Similarity": "score",
		"cu_word":    "CU/wordPos",
	}

	input["customInfo"] = customInfo
	input["customReturn"] = customReturn

	response, err := util.HTTPPostJSON(openAPIURL, input, 10)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
	}
	util.LogTrace.Printf("Raw response from OpenAPI: %s", response)

	DCRet := DCResponse{}
	err = json.Unmarshal([]byte(response), &DCRet)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}

	ret := RetData{}
	if strings.Trim(DCRet.Emotion, " ") != "" {
		splits := strings.Split(DCRet.Emotion, " ")
		ret.Emotion = splits[0]
	}
	if strings.Trim(DCRet.Intent, " ") != "" {
		splits := strings.Split(DCRet.Intent, " ")
		ret.Intent = splits[0]
	}

	if strings.Trim(DCRet.Return, " ") != "" {
		answers := strings.Split(DCRet.Return, "],[")
		ret.Answers = []*string{}
		if len(answers) == 1 {
			answer := strings.Replace(DCRet.Return, "[CMD]:", "", 1)
			ret.Answers = append(ret.Answers, &answer)
		} else {
			lastIdx := len(answers) - 1
			answers[0] = strings.TrimLeft(answers[0], "[")
			answers[lastIdx] = strings.TrimRight(answers[lastIdx], "]")
			for idx := range answers {
				answer := strings.Replace(answers[idx], "[CMD]:", "", 1)
				ret.Answers = append(ret.Answers, &answer)
			}
		}
	} else {
		return nil, ApiError.QA_TEST_FORMAT_ERROR, errors.New("Answer column is empty")
	}
	ret.SimilarQuestion = getSimilarFromCustomReturn(&DCRet.CustomReturn)
	if ret.SimilarQuestion == nil {
		ret.SimilarQuestion = []*QuestionInfo{}
	}
	ret.Tokens = getTokensFromCustomReturn(&DCRet.CustomReturn)
	if ret.Tokens == nil {
		ret.Tokens = []*string{}
	}

	return &ret, ApiError.SUCCESS, nil
}

func DoChatRequestWithOpenAPI(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	openAPIURL := getTestURL()
	if len(openAPIURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	// Prepare for openAPI input
	input := make(map[string]interface{})
	// input["userid"] = user
	// input["appid"] = appid
	// input["text"] = inputData.UserInput
	// input["cmd"] = "chat"

	input["UniqueID"] = "test"
	input["Text1"] = inputData.UserInput
	input["robot"] = appid
	input["UserID"] = user

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

	input["customInfo"] = customInfo
	input["customReturn"] = customReturn

	util.LogTrace.Printf("CustomInfo: %s\n", customInfoStr)
	util.LogTrace.Printf("CustomReturn: %s\n", customReturnStr)

	response, err := util.HTTPPostJSON(openAPIURL, input, 10)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
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
	ret.SimilarQuestion = getSimilarFromCustomReturn(&openAPIRet.CustomReturn)
	if ret.SimilarQuestion == nil {
		ret.SimilarQuestion = []*QuestionInfo{}
	}
	ret.Tokens = getTokensFromCustomReturn(&openAPIRet.CustomReturn)
	if ret.Tokens == nil {
		ret.Tokens = []*string{}
	}

	return &ret, ApiError.SUCCESS, nil
}

func getSimilarFromCustomReturn(customReturn *map[string]interface{}) []*QuestionInfo {
	defer func() {
		if r := recover(); r != nil {
			util.LogInfo.Println("Parse from openapi's customReturn error: ", r)
		}
	}()

	ret := []*QuestionInfo{}

	responseInfo := (*customReturn)["response_other_info"].(map[string]interface{})
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

func getTokensFromCustomReturn(customReturn *map[string]interface{}) []*string {
	defer func() {
		if r := recover(); r != nil {
			util.LogInfo.Println("Parse from openapi's customReturn error: ", r)
		}
	}()

	ret := []*string{}

	tokens := (*customReturn)["cu_word"].([]interface{})
	for _, rawVal := range tokens {
		val := rawVal.(map[string]interface{})
		word := val["word"]
		// pos := val["pos"]
		// temp := fmt.Sprintf("%s/%s", word, pos)
		temp := fmt.Sprintf("%s", word)
		ret = append(ret, &temp)
	}
	return ret
}
