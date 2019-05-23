package FAQ

import (
	"container/list"
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"emotibot.com/emotigo/module/vipshop-admin/imagesManager"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"emotibot.com/emotigo/module/vipshop-admin/websocket"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo     util.ModuleInfo
	RealtimeEvents []websocket.RealtimeEvent
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "faq",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "question/{qid:int}/similar-questions", []string{"edit"}, handleQuerySimilarQuestions),
			util.NewEntryPoint("POST", "question/{qid:string}/similar-questions", []string{"edit"}, handleUpdateSimilarQuestions),

			util.NewEntryPoint("DELETE", "question/{qid:string}/similar-questions", []string{"edit"}, handleDeleteSimilarQuestions),
			util.NewEntryPoint("GET", "questions/similar/search", []string{"view"}, handleSearchSQuestion),
			util.NewEntryPoint("POST", "questions/similar/search", []string{"view"}, handleBatchSearchSQuestion),
			util.NewEntryPoint("POST", "question", []string{"edit"}, handleCreateQuestion),
			util.NewEntryPoint("PUT", "question/{qid:int}", []string{"edit"}, handleUpdateQuestion),
			util.NewEntryPoint("GET", "question/{qid:int}", []string{"view"}, handleQueryQuestion),
			util.NewEntryPoint("GET", "questions/search", []string{"view"}, handleSearchQuestion),
			util.NewEntryPoint("GET", "questions/filter", []string{"view"}, handleQuestionFilter),
			util.NewEntryPoint("POST", "questions/delete", []string{"edit"}, handleDeleteQuestion),
			util.NewEntryPoint("PUT", "questions/commit", []string{"edit"}, handleCommitQuestion),
			util.NewEntryPoint("GET", "RFQuestions", []string{"view"}, handleGetRFQuestions),
			util.NewEntryPoint("POST", "RFQuestions", []string{"edit"}, handleSetRFQuestions),
			util.NewEntryPoint("POST", "RFQuestions/validation", []string{}, handleRFQValidation),
			util.NewEntryPoint("GET", "category/{cid:string}/questions", []string{"view"}, handleCategoryQuestions),
			util.NewEntryPoint("GET", "category/{cid:string}/RFQuestions", []string{}, handleCategoryRFQuestions),
			util.NewEntryPoint("GET", "categories", []string{"view"}, handleGetCategories),
			util.NewEntryPoint("POST", "category/{id:int}", []string{"edit"}, handleUpdateCategories),
			util.NewEntryPoint("PUT", "category", []string{"edit"}, handleAddCategory),
			util.NewEntryPoint("DELETE", "category/{id:int}", []string{"edit"}, handleDeleteCategory),
		},
	}

	RealtimeEvents = []websocket.RealtimeEvent{
		websocket.EventEntrypoint("AddFaq", test),
	}
}

func test(headers websocket.ConnectionHeaders, message []byte) {
	util.LogTrace.Printf("yes, I am a handler, and I got a message: %s", string(message))
}

func handleAddCategory(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	name := ctx.FormValue("categoryname")
	parentID, err := strconv.Atoi(ctx.FormValue("parentid"))
	if err != nil || name == "" {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	parentCategory, err := GetAPICategory(appid, parentID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		return
	}
	lastOperation = time.Now()
	util.LogInfo.Printf("get api category in handleAddCategory took: %s\n", time.Since(lastOperation))

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
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		auditRet = 0
	} else {
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(newCatetory)
	}
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
	util.LogInfo.Printf("add auditLog took in handleAddCategory: %s\n", time.Since(lastOperation))

}

func handleDeleteCategory(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	categoryID, err := ctx.Params().GetInt("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef(err.Error())
	}

	origCategory, err := GetAPICategory(appid, categoryID)
	util.LogInfo.Printf("get api category in handleDeleteCategory took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		return
	} else if origCategory == nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	paths := strings.Split(origCategory.Path, "/")
	path := strings.Join(paths[1:], "/")

	count, err := GetCategoryQuestionCount(appid, origCategory)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		return
	}
	util.LogInfo.Printf("get api category count in handleDeleteCategory took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	err = DeleteAPICategory(appid, origCategory)
	fmtStr := "[%s]:%s，" + util.Msg["DeleteCategoryDesc"]
	auditMessage := fmt.Sprintf(fmtStr, util.Msg["Category"], path, count)
	auditRet := 1
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		auditRet = 0
	} else {
		ctx.StatusCode(http.StatusOK)
	}
	util.LogInfo.Printf("delete category in handleDeleteCategory took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
	util.LogInfo.Printf("write auditlog in handleDeleteCategory took: %s\n", time.Since(lastOperation))
}

func handleUpdateCategories(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	categoryID, err := ctx.Params().GetInt("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef(err.Error())
	}

	newName := ctx.FormValue("categoryname")
	if newName == "" {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	origCategory, err := GetAPICategory(appid, categoryID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		return
	} else if origCategory == nil {
		ctx.StatusCode(http.StatusBadRequest)
	}
	origCategory.Name = newName
	err = UpdateAPICategoryName(appid, categoryID, newName)
	util.LogInfo.Printf("get api category in handleUpdateCategories took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	origPaths := strings.Split(origCategory.Path, "/")
	origPath := strings.Join(origPaths[1:], "/")
	newPath := strings.Join(append(origPaths[1:len(origPaths)-1], newName), "/")

	auditMessage := fmt.Sprintf("[%s]:%s=>%s", util.Msg["Category"], origPath, newPath)
	auditRet := 1
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
		auditRet = 0
	} else {
		ctx.StatusCode(http.StatusOK)
	}
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
	util.LogInfo.Printf("write auditlog in handleUpdateCategories took: %s\n", time.Since(lastOperation))
}

func handleGetCategories(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	categories, err := GetAPICategories(appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
	}
	util.LogInfo.Printf("get api categories in handleGetCategories took: %s\n", time.Since(lastOperation))

	if categories == nil {
		categories = make([]*APICategory, 0)
	}
	ctx.JSON(categories)
}

func handleQuerySimilarQuestions(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	qid, err := ctx.Params().GetInt("qid")
	if err != nil {
		util.LogError.Printf("error while parsing qid, error: %s", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	similarQuestions, err := selectSimilarQuestions(qid, appid)
	if err != nil {
		util.LogError.Printf("error while get similar questions. err: %s\n", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	util.LogInfo.Printf("get similar questions in handleGetCategories took: %s\n", time.Since(lastOperation))

	type Response struct {
		SimilarQuestions []SimilarQuestion `json:"similar_question"`
	}

	response := Response{
		SimilarQuestions: similarQuestions,
	}

	ctx.JSON(response)
}

func handleUpdateSimilarQuestions(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	qid, err := strconv.Atoi(ctx.Params().GetEscape("qid"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	proccessStatus := 0
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	questions, err := selectQuestions([]int{qid}, appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	} else if len(questions) == 0 {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	util.LogInfo.Printf("get questions in handleUpdateSimilarQuestions took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	var question = questions[0]
	questionCategory, err := GetCategory(question.CategoryID, appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("get category in handleUpdateSimilarQuestions took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	categoryName, err := questionCategory.FullName()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	auditMessage := fmt.Sprintf("[相似问题]:[%s][%s]:", categoryName, question.Content)
	// select origin Similarity Questions for audit log
	similarityQuestions, err := selectSimilarQuestions(qid, appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get similar questions in handleUpdateSimilarQuestions took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	var originSimilarityQuestions []string
	for i := 0; i < len(similarityQuestions); i++ {
		originSimilarityQuestions = append(originSimilarityQuestions, similarityQuestions[i].Content)
	}

	body := SimilarQuestionReqBody{}
	if err = ctx.ReadJSON(&body); err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s\n", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	sqs := body.SimilarQuestions

	// update similar questions
	err = updateSimilarQuestions(qid, appid, userID, sqs)
	if err != nil {
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, "更新相似问失败", proccessStatus)
		util.LogError.Println(err)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	util.LogInfo.Printf("update similar questions in handleUpdateSimilarQuestions took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

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
	util.LogInfo.Printf("write audit log in handleUpdateSimilarQuestions took: %s\n", time.Since(lastOperation))
}

func handleDeleteSimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}

// search question by exactly matching content
func handleSearchQuestion(ctx context.Context) {
	lastOperation := time.Now()
	content := strings.TrimSpace(ctx.FormValue("content"))
	util.LogInfo.Printf("check stand question:{%s}", content)
	appid := util.GetAppID(ctx)
	question, err := searchQuestionByContent(content, appid)
	if err == util.ErrSQLRowNotFound {
		squestion, err := searchSQuestionByContent(content, appid)
		if err == util.ErrSQLRowNotFound {
			ctx.StatusCode(http.StatusNotFound)
			return
		} else if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			util.LogError.Printf("searching SQuestion by content [%s] failed, %s", content, err)
			return
		} else {
			ctx.StatusCode(http.StatusOK)
			ctx.JSON(squestion)
		}
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("searching Question by content [%s] failed, %s", content, err)
		return
	} else {
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(question)
	}
	util.LogInfo.Printf("search question content in handleSearchQuestion took: %s\n", time.Since(lastOperation))
}

func handleCategoryQuestions(ctx iris.Context) {
	lastOperation := time.Now()
	cid := ctx.Params().Get("cid")
	appid := util.GetAppID(ctx)
	id, err := strconv.Atoi(cid)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	category, err := GetCategory(id, appid)
	if err == sql.ErrNoRows {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get category in handleCategoryQuestions took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	includeSub := ctx.Request().URL.Query().Get("includeSubCat")
	var categories []Category
	if includeSub == "true" {
		categories, err = category.SubCats()
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			util.LogError.Println(err)
			return
		}

	}
	//Add category itself into total
	categories = append(categories, category)
	questions, err := GetQuestionsByCategories(categories, appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get questions by categories in handleCategoryQuestions took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	ctx.JSON(questions)
}

func handleQuestionFilter(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	condition, err := ParseCondition(ctx)
	if err != nil {
		util.LogError.Printf("Error happened while parsing query options %s\n", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	// fetch question ids and total row number
	qids, aids, err := DoFilter(condition, appid)
	if err != nil {
		util.LogError.Printf("Error happened while Filter questions %s\n", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("filter questions in handleQuestionFilter took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

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

		ctx.JSON(response)
		return
	} else if len(qids) < condition.Limit {
		pagedQIDs = qids
		pagedAIDs = aids
	} else if len(qids) <= start {
		start = len(qids) - condition.Limit
		end = len(qids)
		pagedQIDs = qids[start:end]
		pagedAIDs = aids[start:end]
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
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("fetch questions in handleQuestionFilter took: %s\n", time.Since(lastOperation))

	total := len(qids)
	pageNum := math.Floor(float64(total / condition.Limit))

	response := Response{
		CurPage:     strconv.Itoa(condition.CurPage),
		Questions:   questions,
		PageNum:     pageNum,
		QuestionNum: total,
	}

	ctx.JSON(response)
}

func handleCreateQuestion(ctx context.Context) {
	// 1. create question
	// 2. create answer
	// 3. write audit log
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)

	newQuestion, newAnswers, err := parseQA(ctx)
	if err != nil {
		util.LogError.Printf("Error happened create new question %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	// prepare audit log
	category := Category{
		ID: newQuestion.CategoryId,
	}
	categoryPath, err := category.FullName()
	if err != nil {
		util.LogError.Printf("Error happened create new question %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("get category full name in handleCreateQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	auditMsg := fmt.Sprintf("[标准问题]:[%s]:%s", categoryPath, newQuestion.Content)
	auditRet := 1

	// create question
	qid, err := InsertQuestion(appid, &newQuestion, newAnswers)
	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if !ok || me.Number != 1062 {
			util.LogError.Printf("Error happened create new question %s", err.Error())
			ctx.StatusCode(http.StatusInternalServerError)
		} else {
			util.LogError.Printf("Duplicate Question Content: %s", newQuestion.Content)
			ctx.StatusCode(http.StatusConflict)
		}
		auditRet = 0
	}
	util.LogInfo.Printf("insert question in handleCreateQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	// write audit log
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationAdd, auditMsg, auditRet)
	util.LogInfo.Printf("write auditlog in handleCreateQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	// response
	type Response struct {
		Qid int64 `json:"qid"`
	}
	responseJson := Response{
		Qid: qid,
	}
	ctx.JSON(responseJson)
}

func handleUpdateQuestion(ctx context.Context) {
	// 1. parse question
	// 2. parse answer
	// 3. get old question & answers
	// 4. prepare audit log
	// 5. update question & answers
	// 6. notfiy multicustomer
	// 7. write audit log
	lastOperation := time.Now()
	qid, err := ctx.Params().GetInt("qid")
	if err != nil {
		util.LogError.Printf("Bad question id: %s", err.Error())
		ctx.StatusCode(http.StatusNotFound)
		return
	}

	newQuestion, _, err := parseQA(ctx)
	newQuestion.QuestionId = qid

	appid := util.GetAppID(ctx)

	// get old question & answers
	oldQuestion, err := fetchOldQuestion(appid, qid)
	if err != nil {
		util.LogError.Printf("error while update question: %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("get old question in handleUpdateQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	auditRet := 1
	if err != nil {
		util.LogError.Printf("error while update question: %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	err = UpdateQuestion(appid, &newQuestion)
	if err != nil {
		util.LogError.Printf("error while update quesiton: %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		auditRet = 0
	}
	util.LogInfo.Printf("update question in handleUpdateQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	// write audit log
	err = writeUpdateAuditLog(ctx, auditRet, &oldQuestion, &newQuestion)
	if err != nil {
		util.LogError.Printf("error while update quesiton: %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("write auditlog in handleUpdateQuestion took: %s\n", time.Since(lastOperation))
}

func writeUpdateAuditLog(ctx context.Context, result int, oldQuestion, newQuestion *Question) (err error) {
	// compare if question content is modified
	var auditLog string
	category := Category{
		ID: newQuestion.CategoryId,
	}
	categoryPath, err := category.FullName()
	if err != nil {
		return
	}

	var userID string = util.GetUserID(ctx)
	var userIP string = util.GetUserIP(ctx)

	if oldQuestion.CategoryId != newQuestion.CategoryId {
		// category changed
		oldCategory := Category{
			ID: oldQuestion.CategoryId,
		}
		oldCategoryPath, categoryError := oldCategory.FullName()
		if categoryError != nil {
			err = categoryError
			return
		}
		auditLog = fmt.Sprintf("[标准问题]:[%s][%s]:%s => %s", categoryPath, newQuestion.Content, oldCategoryPath, categoryPath)
		err = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditLog, result)
	}

	if err != nil {
		return
	}

	if oldQuestion.Content != newQuestion.Content {
		auditLog = fmt.Sprintf("[标准问题]：[%s]：%s => %s", categoryPath, oldQuestion.Content, newQuestion.Content)
		err = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditLog, result)
		if err != nil {
			return
		}
	}

	oldAnswersMap := prepareAnswerMap(oldQuestion.Answers)
	newAnswersMap := prepareAnswerMap(newQuestion.Answers)

	var deletedAnswers string
	var addedAnswers string

	for answerID, oldAnswer := range oldAnswersMap {
		newAnswer, ok := newAnswersMap[answerID]
		if !ok {
			// the answer is deleted
			oldAnswerTagStr := answerSliceString(oldAnswer.Dimension, "所有维度")
			if deletedAnswers == "" {
				deletedAnswers = fmt.Sprintf("[标准答案]:[%s][%s]:%s;", categoryPath, newQuestion.Content, oldAnswerTagStr)
			} else {
				deletedAnswers += fmt.Sprintf(" %s;", oldAnswerTagStr)
			}
		} else {
			newAnswerTagStr := answerSliceString(newAnswer.Dimension, "所有维度")
			if !sameStringSlice(oldAnswer.Dimension, newAnswer.Dimension) {
				// answer tags chagned
				// [标准答案]：[分类路径][标准问题][维度]：原维度 => 新维度
				oldAnswerTagStr := answerSliceString(oldAnswer.Dimension, "所有维度")
				auditLog = fmt.Sprintf("[标准答案]:[分类路径][标准问题][维度]:%s => %s", oldAnswerTagStr, newAnswerTagStr)
				err = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditLog, result)
				if err != nil {
					return
				}
			}
			changed, auditLog, anserAuditErr := answerAuditLog(&oldAnswer, &newAnswer)
			if anserAuditErr != nil {
				auditLog = fmt.Sprintf("[标准答案]:写入安全日志失败")
				anserAuditErr = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditLog, result)
				return anserAuditErr
			}
			if changed {
				auditLog = fmt.Sprintf("[标准答案]:[%s][%s][%s]%s", categoryPath, newQuestion.Content, newAnswerTagStr, auditLog)
				err = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditLog, result)
				if err != nil {
					return
				}
			}
		}
	}
	if deletedAnswers != "" {
		err = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationDelete, deletedAnswers, result)
		if err != nil {
			return
		}
	}

	for _, answer := range newQuestion.Answers {
		if answer.AnswerId == 0 {
			newAnswerTagStr := answerSliceString(answer.Dimension, "所有维度")
			// a new added answer
			if addedAnswers == "" {
				addedAnswers = fmt.Sprintf("[标准答案]:[%s][%s][维度]:%s", categoryPath, newQuestion.Content, newAnswerTagStr)
			} else {
				addedAnswers = fmt.Sprintf("%s;%s", addedAnswers, newAnswerTagStr)
			}
		}
	}
	if addedAnswers != "" {
		err = util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationAdd, addedAnswers, result)
	}
	return
}

func prepareAnswerMap(answers []Answer) (answerMap map[int]Answer) {
	answerMap = make(map[int]Answer)
	for _, answer := range answers {
		answerMap[answer.AnswerId] = answer
	}
	return
}

func fetchOldQuestion(appid string, qid int) (question Question, err error) {

	targetQuestion := Question{
		QuestionId: qid,
	}
	targetQuestions := []Question{targetQuestion}

	questions, err := FindQuestions(appid, targetQuestions)
	if err != nil {
		return
	}

	if len(questions) != 1 {
		err = fmt.Errorf("Bad Question Id: %d", qid)
		return
	}

	question = questions[0]

	question.AppID = appid
	question.FetchAnswers()

	for index, _ := range question.Answers {
		answer := &question.Answers[index]
		answer.AppID = appid
		err = answer.Fetch()
		if err != nil {
			return
		}
	}
	return
}

func parseQA(ctx context.Context) (question Question, answers []Answer, err error) {
	// transfrom data struct here
	// QuestionJson => Question
	// AnswerJson => Answer

	questionJson := QuestionJson{}
	err = ctx.ReadJSON(&questionJson)

	question.CategoryId = questionJson.CategoryID
	question.User = questionJson.User
	question.Content = strings.TrimSpace(questionJson.Content)

	appid := util.GetAppID(ctx)
	for _, answerJson := range questionJson.Answers {
		answer := Answer{}
		err = transformaAnswer(appid, &answerJson, &answer)
		if err != nil {
			return
		}
		answers = append(answers, answer)
	}
	question.Answers = answers
	return
}

func transformaAnswer(appid string, answerJson *AnswerJson, answer *Answer) (err error) {
	if answerJson.ID != 0 {
		answer.AnswerId = answerJson.ID
	}

	tagMap, err := LabelMap(appid)
	if err != nil {
		return
	}

	answer.BeginTime = answerJson.BeginTime
	answer.EndTime = answerJson.EndTime
	answer.DynamicMenus = answerJson.DynamicMenu
	answer.RelatedQuestions = answerJson.RelatedQuestions
	answer.AnswerCmd = answerJson.AnswerCMD
	answer.Content = answerJson.Content
	answer.Images = answerJson.Images

	if answer.AnswerCmd == "shopping" {
		answer.AnswerCmdMsg = answerJson.AnswerCMDMsg
	}

	answer.DimensionIDs = answerJson.Dimension
	for _, tagID := range answer.DimensionIDs {
		dimenstion := tagMap[tagID]
		answer.Dimension = append(answer.Dimension, dimenstion)
	}

	if answerJson.NotShow {
		answer.NotShow = 1
	}
	return
}

func answerAuditLog(oldAnswer, newAnswer *Answer) (changed bool, auditLog string, err error) {
	// [标准答案]：[分类路径][标准问题][维度][修改部分]：原内容 => 新内容
	// [标准答案]:[暂无分类][testQ2][WAP端][标准答案]:<p>testA223123fdsafasfdafdsafdsaf</p>=><p>testA223123fdsafasfdafdsafdsafffffff</p>;[生效时间]:永久=>2018-04-19 11:11:00-2018-04-20 11:11:00

	if oldAnswer.Content != newAnswer.Content {
		changed = true
		oldAnswerContent, err := transformAnswerContent(oldAnswer)
		newAnswerContent, err := transformAnswerContent(newAnswer)

		if err != nil {
			auditLog = fmt.Sprintf("%s[标准答案]:%s=>%s;", auditLog, oldAnswer.Content, newAnswer.Content)
			return changed, auditLog, err
		}
		auditLog = fmt.Sprintf("%s[标准答案]:%s=>%s;", auditLog, oldAnswerContent, newAnswerContent)
	}

	if oldAnswer.BeginTime != newAnswer.BeginTime || oldAnswer.EndTime != newAnswer.EndTime {
		changed = true
		var oldDuration string = durationLog(oldAnswer.BeginTime, oldAnswer.EndTime)
		var newDuration string = durationLog(newAnswer.BeginTime, newAnswer.EndTime)
		auditLog = fmt.Sprintf("%s[生效时间]:%s=>%s;", auditLog, oldDuration, newDuration)
	}

	if oldAnswer.NotShow != newAnswer.NotShow {
		changed = true
		auditLog = fmt.Sprintf("%s[不在推荐问内显示]:%s=>%s;", auditLog, notShowLog(oldAnswer.NotShow), notShowLog(newAnswer.NotShow))
	}

	if oldAnswer.AnswerCmd != newAnswer.AnswerCmd {
		changed = true
		auditLog = fmt.Sprintf("%s[指令]:%s=>%s;", auditLog, answerCmdLog(oldAnswer.AnswerCmd), answerCmdLog(newAnswer.AnswerCmd))
	}

	if !sameStringSlice(oldAnswer.DynamicMenus, newAnswer.DynamicMenus) {
		changed = true
		oldDynamicMenuStr := answerSliceString(oldAnswer.DynamicMenus, "无")
		newDynamicMenuStr := answerSliceString(newAnswer.DynamicMenus, "无")
		auditLog = fmt.Sprintf("%s[指定动态菜单]:%s => %s;", auditLog, oldDynamicMenuStr, newDynamicMenuStr)
	}

	if !sameStringSlice(oldAnswer.RelatedQuestions, newAnswer.RelatedQuestions) {
		changed = true
		oldRelatedQuestionsStr := answerSliceString(oldAnswer.RelatedQuestions, "无")
		newRelatedQuestionsStr := answerSliceString(newAnswer.RelatedQuestions, "无")
		auditLog = fmt.Sprintf("%s[指定相关问]:%s => %s;", auditLog, oldRelatedQuestionsStr, newRelatedQuestionsStr)
	}

	return
}

func durationLog(beginTime, endTime string) (duration string) {
	if beginTime == FOREVER_BEGIN && endTime == FOREVER_END {
		duration = "永久"
	} else {
		duration = fmt.Sprintf("%s-%s", beginTime, endTime)
	}
	return
}

func notShowLog(notShow int) (s string) {
	if notShow == 1 {
		s = "是"
	} else {
		s = "否"
	}
	return
}

func sameStringSlice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	s1String := strings.Join(s1, ",")
	s2String := strings.Join(s2, ",")
	return s1String == s2String
}

func answerSliceString(answerLabels []string, defaultStr string) string {
	if len(answerLabels) == 0 {
		return defaultStr
	}
	return strings.Join(answerLabels, ",")
}

func transformAnswerContent(answer *Answer) (transformedContent string, err error) {
	// transform each image url to [<filename>.<extension>]
	var imageIds []interface{}
	for _, id := range answer.Images {
		imageIds = append(imageIds, id)
	}
	metas, err := imagesManager.GetMetaByImageID(imageIds)
	if err != nil {
		return
	}
	transformedContent = answer.Content
	util.LogTrace.Printf("metas: %+v", metas)

	re := regexp.MustCompile("<img.*?/>")
	results := re.FindAllString(answer.Content, -1)
	var rawFileName string
	for _, imaegTag := range results {
		rawFileName = extractRawFileName(imaegTag)
		rawFileName, err = url.QueryUnescape(rawFileName)
		if err != nil {
			// rollback rawFileName
			rawFileName = extractRawFileName(imaegTag)
		}

		util.LogTrace.Printf("rawFileName: %s", rawFileName)
		for _, meta := range metas {
			if rawFileName == meta.RawFileName {
				util.LogTrace.Printf("meta: %+v", meta)
				replaced := fmt.Sprintf("[%s]", meta.FileName)
				transformedContent = strings.Replace(transformedContent, imaegTag, replaced, -1)
			}
		}
	}
	return
}

func extractRawFileName(tag string) string {
	var rawFileName string

	re := regexp.MustCompile("src=\".+?\"")
	results := re.FindAllString(tag, -1)

	for _, imageUrl := range results {
		tokens := strings.Split(imageUrl, "/")
		rawFileName = tokens[len(tokens)-1]
		rawFileName = strings.Replace(rawFileName, "\"", "", -1)
		break
	}
	return rawFileName
}

func handleDeleteQuestion(ctx context.Context) {
	// 1. get to be deleted questions
	// 2. delete questions
	// 2.1 delete squestions
	// 3. write audit log
	lastOperation := time.Now()
	type Parameters struct {
		Qids []int `json:"qids"`
	}

	appid := util.GetAppID(ctx)
	parameters := Parameters{}

	err := ctx.ReadJSON(&parameters)
	if err != nil {
		util.LogError.Printf("Error happened delete questions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	var targetQuestions []Question
	for _, qid := range parameters.Qids {
		question := Question{
			QuestionId: qid,
		}

		targetQuestions = append(targetQuestions, question)
	}

	toBeDeletedQuestions, err := FindQuestions(appid, targetQuestions)
	if err != nil {
		util.LogError.Printf("Error happened delete questions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("get question in handleDeleteQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	auditMsg := ""
	auditRet := 1
	for _, question := range toBeDeletedQuestions {
		auditMsg += fmt.Sprintf("[标准问题]:[%s]:%s", question.CategoryName, question.Content)
	}

	err = DeleteQuestions(appid, targetQuestions)
	if err != nil {
		util.LogError.Printf("Error happened delete questions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		auditRet = 0
	}
	util.LogInfo.Printf("delete questions in handleDeleteQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	//2.1
	err = DeleteSQuestions(appid, targetQuestions)
	if err != nil {
		util.LogError.Printf("Error happened delete squestions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		auditRet = 0
	}
	util.LogInfo.Printf("delete squestions in handleDeleteQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	// write audit log
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationDelete, auditMsg, auditRet)
	util.LogInfo.Printf("write auditlog in handleDeleteQuestion took: %s\n", time.Since(lastOperation))
}

func handleCommitQuestion(ctx context.Context) {
	appid := util.GetAppID(ctx)

	go util.McManualBusiness(appid)
}

func handleQueryQuestion(ctx context.Context) {
	lastOperation := time.Now()
	appid := util.GetAppID(ctx)
	qid, err := ctx.Params().GetInt("qid")
	if err != nil {
		util.LogError.Printf("Error happened while fetching question id, reason: %s", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
	}

	var targetQuestions []Question = []Question{
		Question{
			QuestionId: qid,
		},
	}

	questions, err := FindQuestions(appid, targetQuestions)
	if err != nil {
		util.LogError.Printf("Error happened while get question %d, reason: %s", qid, err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	util.LogInfo.Printf("get questions in handleQueryQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	if len(questions) == 0 {
		util.LogError.Printf("Can not find question: %d", qid)
		ctx.StatusCode(http.StatusNotFound)
		return
	}

	question := questions[0]
	// status -1 means the question was deleted, but not commited
	if question.Status == -1 {
		util.LogInfo.Printf("Target question is deleted: %d\n", qid)
		ctx.StatusCode(http.StatusNotFound)
		return
	}

	question.AppID = appid
	err = question.FetchAnswers()
	if err != nil {
		util.LogError.Printf("Error happened while fetch answers of question %d, reason: %s", qid, err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
	}
	util.LogInfo.Printf("fetch answers of question in handleQueryQuestion took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	for index, _ := range question.Answers {
		answer := &question.Answers[index]
		answer.AppID = appid
		err = answer.Fetch()
		if err != nil {
			return
		}
	}

	ctx.JSON(question)
}

// search simialr question by exactly matching content
func handleSearchSQuestion(ctx context.Context) {
	lastOperation := time.Now()
	content := strings.TrimSpace(ctx.FormValue("content"))
	appid := util.GetAppID(ctx)
	util.LogInfo.Printf("check simalar question {%s}", content)
	squestion, err := searchSQuestionByContent(content, appid)
	if err == util.ErrSQLRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("searching SQuestion by content [%s] failed, %s", content, err)
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(squestion)

	util.LogInfo.Printf("search question content in handleSearchQuestion took: %s\n", time.Since(lastOperation))
}

func handleBatchSearchSQuestion(ctx context.Context) {
	lastOperation := time.Now()

	appid := util.GetAppID(ctx)
	type Parameters struct {
		Question []string `json:"contents"`
	}
	parameters := Parameters{}
	err := ctx.ReadJSON(&parameters)
	util.LogInfo.Printf("%s", parameters.Question)
	if err != nil {
		util.LogError.Printf("Error happened delete questions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	squestion, err := searchBatchSQuestionByContent(parameters.Question, appid)
	if err == util.ErrSQLRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("searching SQuestion by content failed, %s", err)
		return
	}

	l := list.New()
	for i := 0; i < len(parameters.Question); i++ {
		l.PushBack(parameters.Question[i])
	}
	for _, sq := range squestion {
		var next *list.Element
		for e := l.Front(); e != nil; {
			if e.Value.(string) == sq {
				next = e.Next()
				l.Remove(e)
				e = next
			} else {
				e = e.Next()
			}
		}
	}

	var result = []string{}
	for e := l.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(string))
	}
	result = RemoveRepByMap(result)

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(result)

	util.LogInfo.Printf("search question content in handleSearchQuestion took: %s\n", time.Since(lastOperation))
}

func RemoveRepByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}
