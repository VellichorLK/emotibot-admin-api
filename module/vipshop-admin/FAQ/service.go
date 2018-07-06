package FAQ

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sort"
	// "strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func AddAPICategory(appid string, name string, parentID int, level int) (*APICategory, error) {
	newID, err := addApiCategory(appid, name, parentID, level)
	if err != nil {
		return nil, err
	}
	newCategory, err := GetAPICategory(appid, newID)
	if err != nil {
		return nil, err
	}
	return newCategory, nil
}

// sort category by name
type ByName []*APICategory

func (c ByName) Len() int           { return len(c) }
func (c ByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByName) Less(i, j int) bool { return c[i].Name < c[j].Name }

func doSort(categories []*APICategory) {
	if len(categories) != 0 {
		sort.Sort(ByName(categories))
		for _, category := range categories {
			doSort(category.Children)
		}
	}
}

func GetAPICategories(appid string) ([]*APICategory, error) {
	categoriesMap, err := getCategories(appid)
	if err != nil {
		return nil, err
	}

	var ret []*APICategory
	for _, category := range categoriesMap {
		if parent, ok := categoriesMap[category.ParentID]; ok {
			parent.Children = append(parent.Children, category)
		} else if category.ParentID == 0 {
			ret = append(ret, category)
		}
	}
	for _, category := range ret {
		fillCategoryInfo(category, "", 1)
	}
	doSort(ret)

	return ret, nil
}

func GetAPICategory(appid string, categoryID int) (*APICategory, error) {
	categoriesMap, err := getCategories(appid)
	if err != nil {
		return nil, err
	}

	if _, ok := categoriesMap[categoryID]; !ok {
		return nil, nil
	}

	var ret []*APICategory
	for _, category := range categoriesMap {
		if parent, ok := categoriesMap[category.ParentID]; ok {
			parent.Children = append(parent.Children, category)
		} else if category.ParentID == 0 {
			ret = append(ret, category)
		}
	}
	for _, category := range ret {
		fillCategoryInfo(category, "", 1)
	}

	return categoriesMap[categoryID], nil
}

func GetCategoryQuestionCount(appid string, origCategory *APICategory) (int, error) {
	if origCategory == nil {
		return 0, fmt.Errorf("Parameter error")
	}
	// str, _ := json.Marshal(origCategory)
	categoryIDs := getCategoryIDs(origCategory)
	count, err := getQuestionCountInCategories(appid, categoryIDs)
	return count, err
}

func getCategoryIDs(category *APICategory) []int {
	if category == nil {
		return []int{}
	}
	ret := []int{category.ID}
	for _, child := range category.Children {
		ret = append(ret, getCategoryIDs(child)...)
	}
	return ret
}

func DeleteAPICategory(appid string, category *APICategory) error {
	IDs := getCategoryIDs(category)
	db := util.GetMainDB()
	t, err := db.Begin()
	if err != nil {
		return fmt.Errorf("can't aquire transaction lock, %s", err)
	}
	defer t.Commit()

	err = deleteCategories(appid, IDs)
	if err != nil {
		t.Rollback()
		return err
	}
	err = disableQuestionInCategories(appid, IDs)
	if err != nil {
		t.Rollback()
		return err
	}

	return err
}

func UpdateAPICategoryName(appid string, categoryID int, newName string) error {
	err := updateCategoryName(appid, categoryID, newName)
	return err
}

func fillCategoryInfo(category *APICategory, parentPath string, level int) {
	category.Path = fmt.Sprintf("%s/%s", parentPath, category.Name)
	category.Level = level
	for _, child := range category.Children {
		fillCategoryInfo(child, category.Path, level+1)
	}
}

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

	if err = setQuestionDirty(t, qid, appid); err != nil {
		t.Rollback()
		return fmt.Errorf("set Question status failed when update similar questions, error: %s", err)
	}
	t.Commit()

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
		SearchAll:              false,
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
		if page < 0 {
			page = 0
		}
		condition.CurPage = page
	}

	pageLimit, err := strconv.Atoi(limit)
	if err == nil {
		condition.Limit = pageLimit
	}

	return condition, nil
}
