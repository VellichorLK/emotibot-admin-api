package FAQ

import (
	"fmt"
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
		},
	}
}

func handleQuerySimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}

func handleUpdateSimilarQuestions(ctx context.Context) {
	appid := util.GetAppID(ctx)
	qid := ctx.Params().GetEscape("qid")
	proccessStatus := 0
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	qContent, err := selectQuestion(qid, appid)
	if qContent == "" { // return NOT FOUND if question content is empty
		ctx.StatusCode(iris.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	auditMessage := fmt.Sprintf("[标准Q]:%s", qContent)
	// select origin answers for audit log
	originSimilarityAnswers, err := selectSimilarQuestions(qid, appid)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	auditMessage += fmt.Sprintf("[相似问题]:\"%s\" => ", strings.Join(originSimilarityAnswers, "\", \""))
	body := SimilarQuestionReqBody{}
	if err = ctx.ReadJSON(&body); err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}
	sqs := body.SimilarQuestions

	// update similar questions
	err = updateSimilarQuestions(qid, appid, userID, sqs)
	if err != nil {
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, "更新相似问失败", proccessStatus)
		util.LogError.Println(err)
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}
	// save audit log
	var a = make([]string, len(sqs))
	for i, s := range sqs {
		a[i] = s.Content
	}
	proccessStatus = 1
	auditMessage += fmt.Sprintf("\"%s\"", strings.Join(a, "\", \""))
	util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationEdit, auditMessage, proccessStatus)
}

func handleDeleteSimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}
