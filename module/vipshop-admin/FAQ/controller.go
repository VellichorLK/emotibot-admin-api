package FAQ

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "faq",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "question/{qid:string}/similar-questions", []string{"edit"}, handleQuerySimilarQuestions),
			util.NewEntryPoint("POST", "question/{qid:string}/similar-questions", []string{"edit"}, handleUpdateSimilarQuestions),

			util.NewEntryPoint("DELETE", "question/{qid:string}/similar-questions", []string{"edit"}, handleDeleteSimilarQuestions),
			util.NewEntryPoint("POST", "question", []string{"edit"}, handleCreateQuestion),
			util.NewEntryPoint("PUT", "question/{qid:int}", []string{"edit"}, handleUpdateQuestion),
			util.NewEntryPoint("GET", "questions/search", []string{"view"}, handleSearchQuestion),
			util.NewEntryPoint("GET", "questions/filter", []string{"view"}, handleQuestionFilter),
			util.NewEntryPoint("POST", "questions/delete", []string{"edit"}, handleDeleteQuestion),
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
}

func handleAddCategory(ctx context.Context) {
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
}

func handleDeleteCategory(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	categoryID, err := ctx.Params().GetInt("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef(err.Error())
	}

	origCategory, err := GetAPICategory(appid, categoryID)
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
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, auditRet)
}

func handleUpdateCategories(ctx context.Context) {
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
}

func handleGetCategories(ctx context.Context) {
	appid := util.GetAppID(ctx)
	categories, err := GetAPICategories(appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
	}
	ctx.JSON(categories)
}

func handleQuerySimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}

func handleUpdateSimilarQuestions(ctx context.Context) {
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
	var question = questions[0]
	questionCategory, err := GetCategory(question.CategoryID, appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
	}
	categoryName, err := questionCategory.FullName()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
	}
	auditMessage := fmt.Sprintf("[相似问题]:[%s][%s]:", categoryName, question.Content)
	// select origin Similarity Questions for audit log
	originSimilarityQuestions, err := selectSimilarQuestions(qid, appid)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	body := SimilarQuestionReqBody{}
	if err = ctx.ReadJSON(&body); err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
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

func handleDeleteSimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}

// search question by exactly matching content
func handleSearchQuestion(ctx context.Context) {
	content := ctx.FormValue("content")
	appid := util.GetAppID(ctx)
	question, err := searchQuestionByContent(content, appid)
	if err == util.ErrSQLRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("searching Question by content [%s] failed, %s", content, err)
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(question)

}

func handleCategoryQuestions(ctx iris.Context) {
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

	ctx.JSON(questions)
}

func handleQuestionFilter(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	condition, err := ParseCondition(ctx)
	if err != nil {
		util.LogError.Printf("Error happened while parsing query options %s", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	// fetch question ids and total row number
	qids, aids, err := DoFilter(condition, appid)

	if err != nil {
		util.LogError.Printf("Error happened while Filter questions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
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

		ctx.JSON(response)
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

	ctx.JSON(response)
}

func handleCreateQuestion(ctx context.Context) {
	// 1. create question
	// 2. create answer
	// 3. write audit log

	appid := util.GetAppID(ctx)

	questionJson := QuestionJson{}
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
	auditMsg := fmt.Sprintf("[标准问题]:[%s]:%s", categoryPath, questionJson.Content)
	auditRet := 1

	// create question
	qid, err := InsertQuestion(appid, &newQuestion, newAnswers)
	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if !ok || me.Number != 1062 {
			util.LogError.Printf("Error happened create new question %s", err.Error())
			ctx.StatusCode(http.StatusInternalServerError)
		} else {
			util.LogError.Printf("Duplicate Question Content: %s", questionJson.Content)
			ctx.StatusCode(http.StatusConflict)
		}
		auditRet = 0
	}

	// write audit log
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationAdd, auditMsg, auditRet)

	// notify FAQ update
	if err == nil {
		go util.McManualBusiness(appid)
	}

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

	// write audit log
	err = writeUpdateAuditLog(ctx, auditRet, &oldQuestion, &newQuestion)
	if err != nil {
		util.LogError.Printf("error while update quesiton: %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
}

func writeUpdateAuditLog(ctx context.Context, result int,  oldQuestion, newQuestion *Question) (err error) {
	// compare if question content is modified
	var auditLog string
	category := Category {
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
		oldCategory := Category {
			ID: oldQuestion.CategoryId,
		}
		oldCategoryPath, categoryError := oldCategory.FullName()
		if categoryError != nil {
			err = categoryError
			return
		}
		auditLog = fmt.Sprintf("[标准问题]:[%s]:%s => %s", newQuestion.Content, oldCategoryPath, categoryPath)
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
			if deletedAnswers == "" {
				deletedAnswers = fmt.Sprintf("[标准答案]:[%s][%s]:%s;", categoryPath, newQuestion.Content, strings.Join(oldAnswer.Dimension[:], ","))
			} else {
				deletedAnswers += fmt.Sprintf(" %s;", strings.Join(oldAnswer.Dimension[:], ","))
			}
		} else {
			changed, auditLog := answerAuditLog(&oldAnswer, &newAnswer)
			if changed {
				auditLog = fmt.Sprintf("[标准答案]:[%s][%s][%s]%s", categoryPath, newQuestion.Content, strings.Join(newAnswer.Dimension[:], ","), auditLog)
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

	for answerID, newAnswer := range newAnswersMap {
		_, ok := oldAnswersMap[answerID]
		if !ok {
			// a new added answer
			if addedAnswers == "" {
				addedAnswers = fmt.Sprintf("[标准答案]:[%s][%s][维度]:%s ；", categoryPath, newQuestion.Content, strings.Join(newAnswer.Dimension[:], ","))
			} else {
				addedAnswers = fmt.Sprintf("%s %s", addedAnswers, strings.Join(newAnswer.Dimension, ","))
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
	if err !=nil {
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
	question.Content = questionJson.Content

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

func answerAuditLog(oldAnswer, newAnswer *Answer) (changed bool, auditLog string) {
	// [标准答案]：[分类路径][标准问题][维度][修改部分]：原内容 => 新内容
	// [标准答案]:[暂无分类][testQ2][WAP端][标准答案]:<p>testA223123fdsafasfdafdsafdsaf</p>=><p>testA223123fdsafasfdafdsafdsafffffff</p>;[生效时间]:永久=>2018-04-19 11:11:00-2018-04-20 11:11:00

	if oldAnswer.Content != newAnswer.Content {
		changed = true
		auditLog = fmt.Sprintf("%s[标准答案]:%s=>%s;", auditLog, oldAnswer.Content, newAnswer.Content)
	}

	if oldAnswer.BeginTime != newAnswer.BeginTime || oldAnswer.EndTime != newAnswer.EndTime {
		changed = true
		var oldDuration string = durationLog(oldAnswer.BeginTime, oldAnswer.EndTime)
		var newDuration string = durationLog(newAnswer.BeginTime, newAnswer.EndTime)
		auditLog = fmt.Sprintf("%s[生效时间]:%s=>%s", auditLog, oldDuration, newDuration)
	}
	
	if oldAnswer.NotShow != newAnswer.NotShow {
		changed = true
		auditLog = fmt.Sprintf("%s[不在推荐问内显示]:否=>是", auditLog, notShowLog(oldAnswer.NotShow), notShowLog(newAnswer.NotShow))
	}

	if oldAnswer.AnswerCmd != newAnswer.AnswerCmd {
		changed = true
		auditLog = fmt.Sprintf("%s[指令]:%s=>%s", auditLog, answerCmdLog(oldAnswer.AnswerCmd), answerCmdLog(newAnswer.AnswerCmd))
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

func answerCmdLog(cmd string) (chineseCmd string) {
	switch cmd {
	case "":
		chineseCmd = "无指令"
	case "order_track":
		chineseCmd = "物流信息查询"
	case "order_info":
		chineseCmd = "订单信息查询"
	case "scene_id":
		chineseCmd = "场景标识"
	case "cash":
		chineseCmd = "提现"
	case "order_cancel":
		chineseCmd = "取消订单"
	case "apply_for_return":
		chineseCmd = "退货申请"
	case "exchange_goods":
		chineseCmd = "换货申请"
	case "vip_finance":
		chineseCmd = "唯品金融"
	case "query_refund":
		chineseCmd = "查询退款"
	case "shopping":
		chineseCmd = "购物"
	}
	return
}

func handleDeleteQuestion(ctx context.Context) {
	// 1. get to be deleted questions
	// 2. delete questions
	// 3. write audit log

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

	auditMsg := ""
	auditRet := 1
	for _, question := range toBeDeletedQuestions {
		auditMsg += fmt.Sprintf("[标准问题]:[%s]:%s;", question.CategoryName, question.Content)
	}

	err = DeleteQuestions(appid, targetQuestions)
	if err != nil {
		util.LogError.Printf("Error happened delete questions %s", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		auditRet = 0
	}

	// notify FAQ update
	if err == nil {
		go util.McManualBusiness(appid)
	}

	// write audit log
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	util.LogError.Printf("audit message: %s", auditMsg)
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationDelete, auditMsg, auditRet)
}
