package FAQ

import (
	"strings"
	"encoding/json"
	"fmt"
	"strconv"
	// "strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func updateSimilarQuestions(qid int, appid string, user string, sqs []SimilarQuestion) error {
	var err error
	db := util.GetMainDB()
	t, err := db.Begin()
	if err != nil {
		return fmt.Errorf("can't aquire transaction lock, %s", err)
	}
	defer t.Commit()
	// delete old similar questions
	if err = deleteSimilarQuestionsByQuestionID(t, qid, appid); err != nil {
		t.Rollback()
		return fmt.Errorf("delete operation failed, %s", err)
	}

	// put new similar questions
	if err = insertSimilarQuestions(t, qid, appid, user, sqs); err != nil {
		t.Rollback()
		return fmt.Errorf("insert operation failed, %s", err)
	}
	t.Commit()

	// notify multicustomer TODO: update consul directly
	if _, err = util.McManualBusiness(appid); err != nil {
		return fmt.Errorf("error in requesting to MultiCustomer module: %s", err)
	}

	return nil
}

func deleteSimilarQuestions(qid string) error {
	return nil
}

func DoFilter(condition QueryCondition, appid string) ([]int, [][]string, error) {
	qids, aidMap, err := FilterQuestion(condition, appid)
	aids := make([][]string, len(qids))

	if err != nil {
		return qids, aids, err
	}

	for i, qid := range qids {
		aidStr := aidMap[qid]
		aids[i] = strings.Split(aidStr, ",")
	}

	return qids, aids, nil
}

func DoFetch(qids []int, aids [][]string, appid string) ([]Question, error) {
	emptyCondition := QueryCondition{}
	questions, err := FetchQuestions(emptyCondition, qids, aids, appid)

	return questions, err
}

func ParseCondition(param Parameter) (QueryCondition, error) {
	timeSet := param.FormValue("timeset")
	categoryid := param.FormValue("category_id")
	searchStdQ := param.FormValue("search_question")
	searchAns := param.FormValue("search_answer")
	searchDM := param.FormValue("search_dm")
	searchRQ := param.FormValue("search_rq")
	searchAll := param.FormValue("search_all")
	notShowSet := param.FormValue("not_show")
	dimension := param.FormValue("dimension")
	curPage := param.FormValue("cur_page")
	limit := param.FormValue("page_limit")

	var condition = QueryCondition{
		TimeSet:                false,
		BeginTime:              param.FormValue("begin_time"),
		EndTime:                param.FormValue("end_time"),
		Keyword:                param.FormValue("key_word"),
		SearchDynamicMenu:      false,
		SearchRelativeQuestion: false,
		SearchQuestion:         false,
		SearchAnswer:           false,
		SearchAll:				false,
		NotShow:                false,
		CategoryId:             0,
		Limit:                  10,
		CurPage:                0,
	}

	time, _ := strconv.ParseBool(timeSet)
	condition.TimeSet = time

	all, _ := strconv.ParseBool(searchAll)
	condition.SearchAll = all

	question, _ := strconv.ParseBool(searchStdQ)
	condition.SearchQuestion = question

	answer, _ := strconv.ParseBool(searchAns)
	condition.SearchAnswer = answer

	dynamicMenu, _ := strconv.ParseBool(searchDM)
	condition.SearchDynamicMenu = dynamicMenu

	relativeQuestion, _ := strconv.ParseBool(searchRQ)
	condition.SearchRelativeQuestion = relativeQuestion

	notShow, _ := strconv.ParseBool(notShowSet)
	condition.NotShow = notShow

	i, err := strconv.Atoi(categoryid)
	if err != nil {
		return condition, err
	}
	condition.CategoryId = i

	// handle dimension select
	if dimension != "[]" && dimension != "" {
		var dimensionGroups []DimensionGroup
		err := json.Unmarshal([]byte(dimension), &dimensionGroups)
		if err != nil {
			return condition, err
		}
		condition.Dimension = dimensionGroups
	}

	page, err := strconv.Atoi(curPage)
	if err == nil {
		condition.CurPage = page
	}

	pageLimit, err := strconv.Atoi(limit)
	if err == nil {
		condition.Limit = pageLimit
	}

	return condition, nil
}
