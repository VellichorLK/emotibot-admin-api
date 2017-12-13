package FAQ

import (
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
		ModuleName: "FAQ",
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

	util.LogInfo.Printf("qid: %s", qid)
	util.LogInfo.Printf("appid: %s", appid)

	// return NOT FOUND if question id does not exist
	exist, err := findQuestion(qid, appid)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	if !exist {
		ctx.StatusCode(iris.StatusNotFound)
		return
	}

	body := fetchSimilarQuestionList(ctx)
	sqs := body.SimilarQuestions
	user := body.User

	// update similar questions
	updateSimilarQuestions(qid, appid, user, sqs)
}

func fetchSimilarQuestionList(ctx context.Context) *SimilarQuestionReqBody {
	reqBody := &SimilarQuestionReqBody{}
	err := ctx.ReadJSON(reqBody)

	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}

	return reqBody
}

func handleDeleteSimilarQuestions(ctx context.Context) {
	ctx.Writef("[]")
}
