package Robot

import (
	"errors"
	"fmt"
	"encoding/json"
	"strings"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetRobotQuestionCnt(appid string) (int, error) {
	count, err := getAllRobotQACnt(appid)
	if err != nil {
		return 0, err
	}

	return count, err
}

func GetRobotQA(appid string, id int) (*QAInfo, int, error) {
	ret, err := getRobotQA(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return ret, ApiError.SUCCESS, err
}

func GetRobotQAList(appid string) (*RetQAInfo, int, error) {
	list, err := getAllRobotQAList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := RetQAInfo{
		Count: len(list),
		Infos: list,
	}

	return &ret, ApiError.SUCCESS, err
}

func GetRobotQAPage(appid string, page int, listPerPage int) (*RetQAInfo, int, error) {
	list, err := getRobotQAListPage(appid, page, listPerPage)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	count, err := getAllRobotQACnt(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := RetQAInfo{
		Count: count,
		Infos: list,
	}

	return &ret, ApiError.SUCCESS, err
}

func UpdateRobotQA(appid string, id int, info *QAInfo) (int, error) {
	err := updateRobotQA(appid, id, info)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func GetRobotChatInfoList(appid string) ([]*ChatDescription, int, error) {
	ret, err := getRobotChatInfoList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetRobotChat(appid string, id int) (*ChatInfo, int, error) {
	ret, err := getRobotChat(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetRobotChatList(appid string) ([]*ChatInfo, int, error) {
	ret, err := getRobotChatList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetMultiRobotChat(appid string, id []int) ([]*ChatInfo, int, error) {
	ret, err := getMultiRobotChat(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func UpdateMultiChat(appid string, input []*ChatInfoInput) (int, error) {
	if len(input) <= 0 {
		return ApiError.REQUEST_ERROR, errors.New("Invalid request")
	}

	err := updateMultiRobotChat(appid, input)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func GetChatQAList(appid string, keyword string, page int, pageLimit int)(*ChatQAList, int, error) {
	if len(appid) < 0 {
		return nil, ApiError.REQUEST_ERROR, errors.New("Invalid request")
	}

	//produce solr query url
	solrURL := getEnvironment("SOLR_URL")
	if len(solrURL) <= 0 {
		solrURL = "http://172.17.0.1:8081/solr/3rd_core/select?q=(database:appid_robot OR database:chat)"
		util.LogTrace.Printf("ENV \"SOLR_URL\" not exist, use default url")
	}
	solrURL = strings.Replace(solrURL, "appid", appid, -1);
	query := fmt.Sprintf("AND (sentence_original:*%s* OR related_sentences:*%s*)", keyword, keyword)
	
	start, rows := 0, 10 //default query parameter
	if page > 0 && pageLimit > 0 {
		start = (page - 1) * pageLimit
		rows = pageLimit
	}
	url := fmt.Sprintf("%s%s&start=%d&rows=%d&wt=json&indent=true", solrURL, query, start, rows)
	util.LogTrace.Printf("Solr query URL: %s", url)

	//solr query
	content, err := util.HTTPGetSimple(url)
	if err != nil {
		return nil, ApiError.WEB_REQUEST_ERROR, errors.New("Fail to connect to solr")
	}
	
	//parse query result
	solrJson := SolrQueryResponse{}
	err = json.Unmarshal([]byte(content), &solrJson)

	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, errors.New("Fail to parse json")
	}
	if solrJson.ResponseHeader.Status != 0 {
		return nil, ApiError.WEB_REQUEST_ERROR, errors.New("Fail to execute solr query")
	}

	ret := ChatQAList{}
	ret.TotalQACnt = solrJson.Response.NumFound
	ret.ChatQAs = make([]ChatQA, len(solrJson.Response.QAs))
	for qIdx, solrQA := range solrJson.Response.QAs {
		ret.ChatQAs[qIdx].Question = solrQA.Question
		ret.ChatQAs[qIdx].Answers = make([]string, len(solrQA.Answers))
		for aIdx, answer := range solrQA.Answers {
			parsedAnswer := map[string]interface{}{}
			err = json.Unmarshal([]byte(answer), &parsedAnswer)
			if err != nil {
				return nil, ApiError.JSON_PARSE_ERROR, errors.New("Fail to parse json")
			}
			ret.ChatQAs[qIdx].Answers[aIdx] = parsedAnswer["answer"].(string)
		}
	}
	return &ret, ApiError.SUCCESS, nil
}

