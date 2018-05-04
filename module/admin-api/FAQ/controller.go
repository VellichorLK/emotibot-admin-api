package FAQ

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "faq",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "question/{qid}/similar-questions", []string{"edit"}, handleQuerySimilarQuestions),
			util.NewEntryPoint("POST", "question/{qid}/similar-questions", []string{"edit"}, handleUpdateSimilarQuestions),
			util.NewEntryPoint("DELETE", "question/{qid}/similar-questions", []string{"edit"}, handleDeleteSimilarQuestions),
			util.NewEntryPoint("GET", "questions/search", []string{"view"}, handleSearchQuestion),
			util.NewEntryPoint("GET", "questions/filter", []string{"view"}, handleQuestionFilter),

			util.NewEntryPoint("GET", "RFQuestions", []string{"view"}, handleGetRFQuestions),
			util.NewEntryPoint("POST", "RFQuestions", []string{"edit"}, handleSetRFQuestions),

			util.NewEntryPoint("GET", "category/{cid}/questions", []string{"view"}, handleCategoryQuestions),
			util.NewEntryPoint("GET", "categories", []string{"view"}, handleGetCategories),
			util.NewEntryPoint("POST", "category/{id}", []string{"edit"}, handleUpdateCategories),
			util.NewEntryPoint("PUT", "category", []string{"edit"}, handleAddCategory),
			util.NewEntryPoint("DELETE", "category/{id}", []string{"edit"}, handleDeleteCategory),

			util.NewEntryPoint("GET", "labels", []string{"view"}, handleGetLabels),
			util.NewEntryPoint("POST", "label/{id}", []string{"view"}, handleUpdateLabel),
			util.NewEntryPoint("PUT", "label", []string{"view"}, handleAddLabel),
			util.NewEntryPoint("DELETE", "label/{id}", []string{"view"}, handleDeleteLabel),

			util.NewEntryPoint("GET", "rules", []string{"view"}, handleGetRules),
			util.NewEntryPoint("GET", "rule/{id}", []string{"edit"}, handleGetRule),
			util.NewEntryPoint("PUT", "rule/{id}", []string{"edit"}, handleUpdateRule),
			util.NewEntryPoint("POST", "rule", []string{"create"}, handleAddRule),
			util.NewEntryPoint("DELETE", "rule/{id}", []string{"view"}, handleDeleteRule),

			util.NewEntryPoint("GET", "label/{id}/rules", []string{"view"}, handleGetRulesOfLabel),
			util.NewEntryPoint("GET", "rule/{id}/labels", []string{"view"}, handleGetLabelsOfRule),
		},
	}
}

func handleAddCategory(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

	name := r.FormValue("categoryname")
	parentID, err := strconv.Atoi(r.FormValue("parentid"))
	if err != nil || name == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	parentCategory, err := GetAPICategory(appid, parentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var newCatetory *APICategory
	var path string
	if parentCategory == nil {
		newCatetory, err = AddAPICategory(appid, name, 0, 1)
		path = name
	} else {
		newCatetory, err = AddAPICategory(appid, name, parentID, parentCategory.Level+1)
		paths := strings.Split(parentCategory.Path, "/")
		path = strings.Join(append(paths[1:], name), "/")
	}
	auditMessage := fmt.Sprintf("[%s]:%s", util.Msg["Category"], path)
	auditRet := 1
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		auditRet = 0
	} else {
		util.WriteJSON(w, newCatetory)
	}
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
}

func handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)
	categoryID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	origCategory, err := GetAPICategory(appid, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if origCategory == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	paths := strings.Split(origCategory.Path, "/")
	path := strings.Join(paths[1:], "/")

	count, err := GetCategoryQuestionCount(appid, origCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = DeleteAPICategory(appid, origCategory)
	fmtStr := "[%s]:%s，" + util.Msg["DeleteCategoryDesc"]
	auditMessage := fmt.Sprintf(fmtStr, util.Msg["Category"], path, count)
	auditRet := 1
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		auditRet = 0
	}
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
	util.ConsulUpdateFAQ(appid)
}

func handleUpdateCategories(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)
	categoryID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	newName := r.FormValue("categoryname")
	if newName == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	origCategory, err := GetAPICategory(appid, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if origCategory == nil {
		http.Error(w, "", http.StatusBadRequest)
	}
	origCategory.Name = newName
	err = UpdateAPICategoryName(appid, categoryID, newName)

	origPaths := strings.Split(origCategory.Path, "/")
	origPath := strings.Join(origPaths[1:], "/")
	newPath := strings.Join(append(origPaths[1:len(origPaths)-1], newName), "/")

	auditMessage := fmt.Sprintf("[%s]:%s=>%s", util.Msg["Category"], origPath, newPath)
	auditRet := 1
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		auditRet = 0
	}
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
}

func handleGetCategories(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	categories, err := GetAPICategories(appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	util.WriteJSON(w, categories)
}

func handleQuerySimilarQuestions(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func handleUpdateSimilarQuestions(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "qid")
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	proccessStatus := 0
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

	questions, err := selectQuestions([]int{qid}, appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(questions) == 0 {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	var question = questions[0]
	questionCategory, err := GetCategory(question.CategoryID, appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categoryName, err := questionCategory.FullName(appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	auditMessage := fmt.Sprintf("[相似问题]:[%s][%s]:", categoryName, question.Content)
	// select origin Similarity Questions for audit log
	originSimilarityQuestions, err := selectSimilarQuestions(qid, appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body := SimilarQuestionReqBody{}
	if err = util.ReadJSON(r, &body); err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	sqs := body.SimilarQuestions

	// update similar questions
	err = updateSimilarQuestions(qid, appid, userID, sqs)
	if err != nil {
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, "更新相似问失败", proccessStatus)
		util.LogError.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	//sqsStr 移除了沒更動的相似問
	var sqsStr []string
	//contentMatching 邏輯: 移除掉一模一樣的新舊相似問內容, 來寫audit log
contentMatching:
	for i := 0; i < len(sqs); i++ {
		sq := sqs[i].Content
		for j := len(originSimilarityQuestions) - 1; j >= 0; j-- {
			oldSq := originSimilarityQuestions[j]
			if sq == oldSq {
				originSimilarityQuestions = append(originSimilarityQuestions[:j], originSimilarityQuestions[j+1:]...)
				continue contentMatching
			}
		}
		sqsStr = append(sqsStr, sq)
	}
	sort.Strings(originSimilarityQuestions)
	sort.Strings(sqsStr)

	proccessStatus = 1
	operation := util.AuditOperationEdit
	//當全部都是新的(原始的被扣完)行為要改成新增, 全部都是舊的(新的是空的)行為要改成刪除
	if len(originSimilarityQuestions) == 0 {
		operation = util.AuditOperationAdd
		auditMessage += fmt.Sprintf("%s", strings.Join(sqsStr, ";"))
	} else if len(sqsStr) == 0 {
		operation = util.AuditOperationDelete
		auditMessage += fmt.Sprintf("%s", strings.Join(originSimilarityQuestions, ";"))
	} else {
		auditMessage += fmt.Sprintf("%s=>%s", strings.Join(originSimilarityQuestions, ";"), strings.Join(sqsStr, ";"))

	}
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, operation, auditMessage, proccessStatus)

}

func handleDeleteSimilarQuestions(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// search question by exactly matching content
func handleSearchQuestion(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("content")
	appid := util.GetAppID(r)
	question, err := searchQuestionByContent(content, appid)
	if err == util.ErrSQLRowNotFound {
		http.Error(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		util.LogError.Printf("searching Question by content [%s] failed, %s", content, err)
		return
	}
	util.WriteJSON(w, question)
}

//Retrun JSON Formatted RFQuestion array, if question is invalid, id & categoryId will be 0
func handleGetRFQuestions(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	questions, err := GetRFQuestions(appid)
	if err != nil {
		util.LogError.Printf("Get RFQuestions failed, %v\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, questions)
}

func handleSetRFQuestions(w http.ResponseWriter, r *http.Request) {
	var args UpdateRFQuestionsArgs
	appid := util.GetAppID(r)
	err := util.ReadJSON(r, &args)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if err = SetRFQuestions(args.Contents, appid); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}

}

func handleCategoryQuestions(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "cid")
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	category, err := GetCategory(id, appid)
	if err == sql.ErrNoRows {
		http.Error(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	includeSub := r.URL.Query().Get("includeSubCat")
	var categories []Category
	if includeSub == "true" {
		categories, err = category.SubCats()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			util.LogError.Println(err)
			return
		}

	}
	//Add category itself into total
	categories = append(categories, category)
	questions, err := GetQuestionsByCategories(categories, appid)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}

	util.WriteJSON(w, questions)
}

func handleQuestionFilter(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	// parse QueryCondition
	condition, err := ParseCondition(r)
	if err != nil {
		util.LogError.Printf("Error happened while parsing query options %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	// fetch question ids and total row number
	qids, aids, err := DoFilter(condition, appid)

	if err != nil {
		util.LogError.Printf("Error happened while Filter questions %s", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// paging qids
	start := condition.CurPage * condition.Limit
	end := start + condition.Limit

	// fetch returned question and answers
	type Response struct {
		CurPage     string     `json:"CurPage"`
		Questions   []Question `json:"QueryResult"`
		PageNum     float64    `json:"TotalNum"`
		QuestionNum int        `json:"TotalQuestionNum"`
	}

	var pagedQIDs []int
	var pagedAIDs [][]string
	if len(qids) == 0 {
		response := Response{
			CurPage:     "0",
			Questions:   make([]Question, 0),
			PageNum:     0,
			QuestionNum: 0,
		}

		util.WriteJSON(w, response)
		return
	} else if len(qids) < condition.Limit {
		pagedQIDs = qids
		pagedAIDs = aids
	} else if len(qids) < end {
		end = len(qids)
		pagedQIDs = qids[start:end]
		pagedAIDs = aids[start:end]
	} else {
		pagedQIDs = qids[start:end]
		pagedAIDs = aids[start:end]
	}

	questions, err := DoFetch(pagedQIDs, pagedAIDs, appid)
	if err != nil {
		util.LogError.Printf("Error happened Fetch questions %s", err.Error())
	}

	total := len(qids)
	pageNum := math.Floor(float64(total / condition.Limit))

	response := Response{
		CurPage:     strconv.Itoa(condition.CurPage),
		Questions:   questions,
		PageNum:     pageNum,
		QuestionNum: total,
	}

	util.WriteJSON(w, response)
}
