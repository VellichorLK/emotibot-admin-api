package QA

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/FAQ"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func DoChatRequestWithBFOPController(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	controllerURL := getControllerURL()
	if len(controllerURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	input := make(map[string]interface{})
	input["uniqueId"] = genRandomUUIDSameAsOpenAPI()
	input["question"] = inputData.UserInput
	input["robotId"] = appid
	input["userId"] = user

	input["customInfo"] = inputData.Info
	if input["customInfo"] == nil {
		input["customInfo"] = map[string]string{}
	}

	customHeader := make(map[string]string)
	customHeader["X-Lb-Uid"] = user

	// for openapi
	customHeader["uuid"] = genRandomUUIDSameAsOpenAPI()
	customHeader["appId"] = appid
	customHeader["userId"] = user
	input["text"] = inputData.UserInput
	input["extendData"] = inputData.ExtendData

	response, err := util.HTTPPostJSONWithHeader(controllerURL, input, 10, customHeader)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
	}
	logger.Trace.Printf("Raw response from Controller: %s", response)

	controllerRet := BFOPControllerResponse{}
	err = json.Unmarshal([]byte(response), &controllerRet)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}

	ret := RetData{}
	ret.Answers = make([]*string, len(controllerRet.Answer))
	for idx, controllerAns := range controllerRet.Answer {
		temp, err := json.Marshal(controllerAns)
		if err != nil || string(temp) == "" {
			continue
		}
		str := string(temp)
		ret.Answers[idx] = &str
	}

	if controllerRet.Info != nil {
		ret.Emotion = controllerRet.Info.Emotion
		ret.Intent = controllerRet.Info.Intent
		ret.Tokens = controllerRet.Info.Tokens
		ret.Module = controllerRet.Info.Module
		ret.TaskInfo = controllerRet.Info.TaskInfo
	}
	ret.OpenAPIReturn = controllerRet.Status

	return &ret, ApiError.SUCCESS, nil
}

func DoChatRequestWithController(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	controllerURL := getControllerURL()
	if len(controllerURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	input := make(map[string]interface{})
	input["uniqueId"] = genRandomUUIDSameAsOpenAPI()
	input["question"] = inputData.UserInput
	input["robotId"] = appid
	input["userId"] = user

	customInfo := map[string]string{
		"platform": inputData.Platform,
		"brand":    inputData.Brand,
		"sex":      inputData.Gender,
	}
	input["customInfo"] = customInfo

	customHeader := make(map[string]string)
	customHeader["X-Lb-Uid"] = user

	response, err := util.HTTPPostJSONWithHeader(controllerURL, input, 10, customHeader)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
	}
	logger.Trace.Printf("Raw response from Controller: %s", response)

	controllerRet := ControllerResponse{}
	err = json.Unmarshal([]byte(response), &controllerRet)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}

	ret := RetData{}
	ret.Emotion = controllerRet.Emotion
	ret.Intent = controllerRet.Intent

	ret.Answers = make([]*string, len(controllerRet.Answer))
	for idx, controllerAns := range controllerRet.Answer {
		temp, err := json.Marshal(controllerAns)
		if err != nil || string(temp) == "" {
			continue
		}
		str := string(temp)
		ret.Answers[idx] = &str
	}

	ret.Tokens = controllerRet.Tokens
	ret.SimilarQuestion = make([]*QuestionInfo, len(controllerRet.RelatedQuestion))
	for idx, q := range controllerRet.RelatedQuestion {
		ret.SimilarQuestion[idx] = &QuestionInfo{inputData.UserInput, q.Question, q.Score}
	}
	ret.OpenAPIReturn = controllerRet.Status
	ret.Tokens = controllerRet.Tokens

	return &ret, ApiError.SUCCESS, nil
}

func DoChatRequestWithDC(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	openAPIURL := getTestURL()
	if len(openAPIURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	// Prepare for DC input
	input := make(map[string]interface{})
	input["UniqueID"] = genRandomUUIDSameAsOpenAPI()
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

	customHeader := make(map[string]string)
	customHeader["X-Lb-Uid"] = user

	response, err := util.HTTPPostJSONWithHeader(openAPIURL, input, 10, customHeader)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
	}
	logger.Trace.Printf("Raw response from OpenAPI: %s", response)

	DCRet := DCResponse{}
	err = json.Unmarshal([]byte(response), &DCRet)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}

	ret := RetData{}
	cuRet := getFirstCUData(DCRet.Emotion)
	if cuRet != nil {
		ret.Emotion = cuRet.Item
	}
	cuRet = getFirstCUData(DCRet.Intent)
	if cuRet != nil {
		ret.Intent = cuRet.Item
	}

	// Parse for multi answer in format [ans1],[ans2],[[CMD]:{something}]
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

	ret.SimilarQuestion = getSimilarFromDCCustomReturn(&DCRet.CustomReturn)
	if ret.SimilarQuestion == nil {
		ret.SimilarQuestion = []*QuestionInfo{}
	}
	ret.Tokens = getTokensFromDCCustomReturn(&DCRet.CustomReturn)
	if ret.Tokens == nil {
		ret.Tokens = []*string{}
	}

	return &ret, ApiError.SUCCESS, nil
}

func getFirstCUData(input interface{}) *CUDataFromDC {
	defer func() {
		if r := recover(); r != nil {
			logger.Error.Printf("Parse cudata error from %#v", input)
			return
		}
	}()

	jsonByte, _ := json.Marshal(input)
	ret := CUDataFromDC{}

	switch t := input.(type) {
	case map[string]interface{}:
		logger.Trace.Printf("Parse %v for hash", t)
		json.Unmarshal(jsonByte, &ret)
		return &ret
	case []interface{}:
		logger.Trace.Printf("Parse %v for array", t)
		temp := []*CUDataFromDC{}
		json.Unmarshal(jsonByte, &temp)
		if len(temp) > 0 {
			return temp[0]
		}
		return nil
	}
	return nil
}

func DoChatRequestWithOpenAPI(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	logger.Trace.Println("Chat with OPENAPI controller")
	openAPIURL := getTestURL()
	if len(openAPIURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	// Prepare for openAPI input
	input := make(map[string]interface{})
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

	input["customInfo"] = customInfo
	input["customReturn"] = customReturn

	logger.Trace.Printf("CustomInfo: %s\n", customInfoStr)
	logger.Trace.Printf("CustomReturn: %s\n", customReturnStr)

	logger.Trace.Printf("Call OpenAPI With: %+v\n", input)
	response, err := util.HTTPPostJSONWithHeader(openAPIURL, input, 10, map[string]string{
		"appId":  appid,
		"userId": user,
	})
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
	}
	logger.Trace.Printf("Raw response from OpenAPI: %s", response)

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
	ret.SimilarQuestion = getSimilarFromOpenAPICustomReturn(&openAPIRet.CustomReturn)
	if ret.SimilarQuestion == nil {
		ret.SimilarQuestion = []*QuestionInfo{}
	}
	ret.Tokens = getTokensFromOpenAPICustomReturn(&openAPIRet.CustomReturn)
	if ret.Tokens == nil {
		ret.Tokens = []*string{}
	}

	return &ret, ApiError.SUCCESS, nil
}

func DoChatRequestWithBFOPOpenAPI(appid string, user string, inputData *QATestInput) (*RetData, int, error) {
	logger.Trace.Println("Chat with BFOPOPENAPI controller")
	openAPIURL := getTestURL()
	if len(openAPIURL) == 0 {
		return nil, ApiError.REQUEST_ERROR, nil
	}

	// Prepare for openAPI input
	input := make(map[string]interface{})
	header := map[string]string{}

	header["userid"] = user
	header["appid"] = appid
	header["uuid"] = genRandomUUIDSameAsOpenAPI()

	input["text"] = inputData.UserInput
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

	input["customInfo"] = customInfo
	extendData := map[string]bool{
		"online": true,
	}
	input["extend_data"] = extendData

	logger.Trace.Printf("CustomInfo: %s\n", customInfoStr)

	logger.Trace.Printf("Call OpenAPI With: %+v\n", input)
	response, err := util.HTTPPostJSONWithHeader(openAPIURL, input, 10, header)
	if err != nil {
		return nil, ApiError.OPENAPI_URL_ERROR, err
	}
	logger.Trace.Printf("Raw response from OpenAPI: %s\n", response)

	openAPIRet := BFOPOpenAPIResponse{}
	err = json.Unmarshal([]byte(response), &openAPIRet)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}

	ret := RetData{}
	if openAPIRet.Info != nil {
		ret.Emotion = openAPIRet.Info.Emotion
		ret.Intent = openAPIRet.Info.Intent
		ret.Tokens = openAPIRet.Info.Tokens
	}

	if openAPIRet.Data != nil && len(openAPIRet.Data) > 0 {
		for idx := range openAPIRet.Data {
			temp, err := json.Marshal(openAPIRet.Data[idx])
			if err != nil || string(temp) == "" {
				logger.Trace.Println("Marshal to json fail:", err.Error())
				continue
			}
			str := string(temp)
			ret.Answers = append(ret.Answers, &str)
		}
	} else {
		return nil, ApiError.QA_TEST_FORMAT_ERROR, errors.New("Answer column is empty")
	}

	ret.OpenAPIReturn = openAPIRet.Status
	return &ret, ApiError.SUCCESS, nil
}

func getSimilarFromOpenAPICustomReturn(customReturn *map[string]interface{}) []*QuestionInfo {
	defer func() {
		if r := recover(); r != nil {
			logger.Info.Println("Parse from openapi's customReturn error: ", r)
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

func getTokensFromOpenAPICustomReturn(customReturn *map[string]interface{}) []*string {
	defer func() {
		if r := recover(); r != nil {
			logger.Info.Println("Parse from openapi's customReturn error: ", r)
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

func getSimilarFromDCCustomReturn(customReturn *map[string]interface{}) []*QuestionInfo {
	return getSimilarFromOpenAPICustomReturn(customReturn)
}

func getTokensFromDCCustomReturn(customReturn *map[string]interface{}) []*string {
	defer func() {
		if r := recover(); r != nil {
			logger.Info.Println("Parse from openapi's customReturn error: ", r)
		}
	}()

	ret := []*string{}

	// real token may in customReturn.response_other_info.token or customReturn.cu_word
	cuWord, ok := (*customReturn)["cu_word"]
	var tokens []interface{}
	if ok {
		tokens = cuWord.([]interface{})
		for _, rawVal := range tokens {
			val := rawVal.(map[string]interface{})
			word := val["word"]
			pos := val["pos"]
			temp := fmt.Sprintf("%s/%s", word, pos)
			ret = append(ret, &temp)
		}
		return ret
	}

	responseOtherInfo, ok := (*customReturn)["response_other_info"].(map[string]interface{})
	if ok {
		tokens = responseOtherInfo["token"].([]interface{})
	}

	for _, rawVal := range tokens {
		val := rawVal.(string)
		ret = append(ret, &val)
	}
	return ret
}

func genRandomUUIDSameAsOpenAPI() string {
	now := time.Now()
	randomNum := rand.Intn(900) + 100
	ret := fmt.Sprintf("%d%02d%02d%02d%02d%02d%06d%03d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000,
		randomNum)
	return ret
}

func genQAExportAuditLog(appid string, condition *FAQ.QueryCondition, taskID int) (string, error) {
	var content string
	if FAQ.HasCondition(*condition) {
		// rule
		// [部分导出]：[生效时间][全部/问题/答案/指定动态菜单/指定相关问/不在推荐问内显示：关键字][维度][分类路径]：文件名
		var timeCondition string
		var keywordCondition string
		var dimensionCondition string
		var categoryCondition string

		if condition.TimeSet {
			timeCondition = "[时间段：%s%s%s%s%s-%s%s%s%s%s]"
			startTimeSlices, err := parseDatetimeStr(condition.BeginTime)
			endTimeSlices, err := parseDatetimeStr(condition.EndTime)

			if err != nil {
				return "", err
			}

			timeCondition = fmt.Sprintf(
				timeCondition,
				startTimeSlices[0],
				startTimeSlices[1],
				startTimeSlices[2],
				startTimeSlices[3],
				startTimeSlices[4],
				endTimeSlices[0],
				endTimeSlices[1],
				endTimeSlices[2],
				endTimeSlices[3],
				endTimeSlices[4],
			)
		}

		keywordType := parseKeywordType(condition)
		if keywordType != "" {
			keywordCondition = fmt.Sprintf("[%s：%s]", keywordType, condition.Keyword)
		}

		if len(condition.Dimension) > 0 {
			dimensionStr := genDimensionStr(condition)
			dimensionCondition = fmt.Sprintf("[维度：%s]", dimensionStr)
		}

		categoryStr, err := genCategoryStr(appid, condition)
		if err != nil {
			return content, err
		}
		categoryCondition = fmt.Sprintf("[%s]", categoryStr)

		content = "[部分导出]：%s%s%s%s：other_%d.xlsx"
		content = fmt.Sprintf(content, timeCondition, keywordCondition, dimensionCondition, categoryCondition, taskID)
	} else {
		content = fmt.Sprintf("[全量导出]：other_%d.xlsx", taskID)
	}
	return content, nil
}

func parseDatetimeStr(datetime string) ([]string, error) {
	// expect format yyyy-MM-dd hh:mm:ss
	var dateAndTime []string = make([]string, 5)
	datetimeSlice := strings.Split(datetime, " ")

	if len(datetimeSlice) != 2 {
		return dateAndTime, fmt.Errorf("datetime format incorrect")
	}

	dateString := datetimeSlice[0]
	timeString := datetimeSlice[1]

	dateSlice := strings.Split(dateString, "-")
	if len(dateSlice) != 3 {
		return dateAndTime, fmt.Errorf("datetime format incorrect 2")
	}
	dateAndTime[0] = dateSlice[0]
	dateAndTime[1] = dateSlice[1]
	dateAndTime[2] = dateSlice[2]

	timeSlice := strings.Split(timeString, ":")
	if len(dateSlice) != 3 {
		return dateAndTime, fmt.Errorf("datetime format incorrect 3")
	}
	dateAndTime[3] = timeSlice[0]
	dateAndTime[4] = timeSlice[1]

	return dateAndTime, nil
}

func parseKeywordType(condition *FAQ.QueryCondition) string {
	var keywordType string

	if condition.SearchAll && condition.Keyword != "" {
		keywordType = "全部"
	} else if condition.SearchQuestion && condition.Keyword != "" {
		keywordType = "问题"
	} else if condition.SearchAnswer && condition.Keyword != "" {
		keywordType = "答案"
	} else if condition.SearchDynamicMenu {
		keywordType = "指定动态菜单"
	} else if condition.SearchRelativeQuestion {
		keywordType = "指定相关问"
	} else if condition.NotShow {
		keywordType = "不在推荐问内显示"
	}

	return keywordType
}

func genDimensionStr(condition *FAQ.QueryCondition) string {
	var dimensionStr bytes.Buffer
	for index, dimension := range condition.Dimension {
		cleanContent := strings.Replace(dimension.Content, "#", "", -1)
		if index == 0 {
			dimensionStr.WriteString(cleanContent)
		} else {
			dimensionStr.WriteString(",")
			dimensionStr.WriteString(cleanContent)
		}
	}

	return dimensionStr.String()
}

func genCategoryStr(appid string, condition *FAQ.QueryCondition) (string, error) {
	var categoryPath string
	var err error
	switch condition.CategoryId {
	case 0:
		categoryPath = "全部分类"
	case -1:
		categoryPath = "暂无分类"
	default:
		category := FAQ.Category{
			ID: condition.CategoryId,
		}

		categoryPath, err = category.FullName(appid)
		if err != nil {
			fmt.Println("error: ", err.Error())
			return categoryPath, err
		}
		categorySlice := strings.Split(categoryPath, "/")

		if len(categorySlice) > 2 {
			categoryPath = fmt.Sprintf("%s/%s", categorySlice[0], categorySlice[1])
		} else {
			categoryPath = categorySlice[0]
		}
	}

	return categoryPath, err
}
