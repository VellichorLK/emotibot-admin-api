package FAQ

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"

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
			util.NewEntryPoint("GET", "questions/search", []string{"view"}, handleSearchQuestion),
			util.NewEntryPoint("GET", "questions/filter", []string{"view"}, handleQuestionFilter),
		},
	}
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

	question, err := selectQuestion(qid, appid)
	if err == sql.ErrNoRows {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	questionCategory, err := GetCategoryFullPath(question.CategoryID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
	}
	auditMessage := fmt.Sprintf("[相似問題]:[%s][%s]:", questionCategory, question.Content)
	// select origin answers for audit log
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
auditLogging:
	for i := len(sqs) - 1; i >= 0; i-- {
		sq := sqs[i].Content
		for j := len(originSimilarityQuestions) - 1; j >= 0; j-- {
			oldSq := originSimilarityQuestions[j]
			if sq == oldSq {
				originSimilarityQuestions = append(originSimilarityQuestions[:j], originSimilarityQuestions[j+1:]...)
				continue auditLogging
			}
		}
		sqsStr = append(sqsStr, sq)
	}
	sort.Strings(originSimilarityQuestions)
	sort.Strings(sqsStr)
	auditMessage += fmt.Sprintf("%s=>%s", strings.Join(originSimilarityQuestions, ";"), strings.Join(sqsStr, ";"))

	proccessStatus = 1
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, proccessStatus)
}

func handleDeleteSimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}

// search question by exactly matching content
func handleSearchQuestion(ctx context.Context) {
	content := ctx.FormValue("content")
	question, err := searchQuestionByContent(content)
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
	qids, _, err := DoFilter(condition, appid)
	if err != nil {
		util.LogError.Printf("Error happened while Filter questions %s", err.Error())
	}

	// paging qids
	start := condition.CurPage * condition.Limit
	end := start + condition.Limit

	// fetch returned question and answers
	questions, err := FetchQuestions(condition, qids[start:end], "vipshop")
	if err != nil {
		util.LogError.Printf("Error happened Fetch questions %s", err.Error())
	}


	util.LogError.Printf("Error happened Fetch questions %v", questions)
	ctx.JSON(questions)
}
