package FAQ

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

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
			util.NewEntryPoint("GET", "questions/search", []string{"view"}, handleSearchQuestion),
			util.NewEntryPoint("GET", "RFQuestion", []string{"view"}, handleGetRFQuestions),
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
	questionCategory, err := GetCategory(question.CategoryID)
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

func handleGetRFQuestions(ctx iris.Context) {
	questions, err := GetRFQuestions()
	if err != nil {
		util.LogError.Printf("Get RFQuestions failed, %v\n", err)
	}
	ctx.JSON(questions)
}

func HandleSetRFQuestions(ctx iris.Context) {
	value := ctx.Request().URL.Query()
	if _, ok := value["id"]; !ok {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	var groupID = make([]int, len(value["id"]))
	for i, id := range value["id"] {
		var err error
		groupID[i], err = strconv.Atoi(id)
		if err != nil {
			util.LogError.Printf("id %s parse failed, %v \n", id, err)
			ctx.StatusCode(http.StatusBadRequest)
			return
		}
	}
	appID := util.GetAppID(ctx)
	questions, err := selectQuestions(groupID, appID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	if len(questions) == 0 {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	var contents = make([]string, len(questions))
	for i, q := range questions {
		contents[i] = q.Content
	}
	if err = InsertRFQuestions(contents); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
}

func handleCategoryQuestions(ctx iris.Context) {
	id, err := ctx.Params().GetInt("cid")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	category, err := GetCategory(id)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	includeSub, err := ctx.Params().GetBool("")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	var categories []Category
	if includeSub {
		categories, err = category.SubCats()
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			util.LogError.Println(err)
			return
		}

	}
	//Add category itself into total
	categories = append(categories, category)
	questions, err := GetQuestionsByCategories(categories)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}

	ctx.JSON(questions)
}
